package app.pwhs.blockads.desktop

import com.sun.jna.Memory
import com.sun.jna.Native
import com.sun.jna.Pointer
import com.sun.jna.WString
import com.sun.jna.platform.win32.COM.COMUtils
import com.sun.jna.platform.win32.COM.Unknown
import com.sun.jna.platform.win32.Guid
import com.sun.jna.platform.win32.Ole32
import com.sun.jna.platform.win32.WinDef
import com.sun.jna.platform.win32.WinNT
import com.sun.jna.ptr.IntByReference
import com.sun.jna.ptr.PointerByReference
import com.sun.jna.win32.StdCallLibrary
import java.nio.file.Path

object JumpListManager {
    private val clsidDestinationList = Guid.GUID("{77f10cf0-3db5-4966-b520-b7c54fd35ed6}")
    private val iidCustomDestinationList = Guid.GUID("{6332debf-87b5-4670-90c0-5e57b408a49e}")
    private val clsidEnumerableObjectCollection = Guid.GUID("{2D3468C1-36A7-43B6-AC24-D3F02FD9607A}")
    private val iidObjectCollection = Guid.GUID("{5632b1a4-e38a-400a-928a-d4cd63230295}")
    private val iidObjectArray = Guid.IID("{92CA9DCD-5622-4bba-A805-5E9F541BD8C9}")
    private val clsidShellLink = Guid.GUID("{00021401-0000-0000-C000-000000000046}")
    private val iidShellLinkW = Guid.GUID("{000214F9-0000-0000-C000-000000000046}")
    private val iidPropertyStore = Guid.IID("{886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99}")

    fun install() {
        if (!System.getProperty("os.name", "").startsWith("Windows", ignoreCase = true)) return
        val executable = applicationExecutable()
        val hr = Ole32.INSTANCE.CoInitializeEx(Pointer.NULL, Ole32.COINIT_APARTMENTTHREADED)
        val initialized = hr.toInt() >= 0
        if (!initialized && hr.toInt() != 0x80010106.toInt()) COMUtils.checkRC(hr)
        try {
            val destinationRef = PointerByReference()
            COMUtils.checkRC(
                Ole32.INSTANCE.CoCreateInstance(
                    clsidDestinationList,
                    Pointer.NULL,
                    1,
                    iidCustomDestinationList,
                    destinationRef,
                )
            )
            val destination = CustomDestinationList(destinationRef.value)
            try {
                val slots = IntByReference()
                val removedRef = PointerByReference()
                COMUtils.checkRC(destination.beginList(slots, Guid.REFIID(iidObjectArray), removedRef))
                if (removedRef.value != null && removedRef.value != Pointer.NULL) Unknown(removedRef.value).Release()

                val collectionRef = PointerByReference()
                COMUtils.checkRC(
                    Ole32.INSTANCE.CoCreateInstance(
                        clsidEnumerableObjectCollection,
                        Pointer.NULL,
                        1,
                        iidObjectCollection,
                        collectionRef,
                    )
                )
                val collection = ObjectCollection(collectionRef.value)
                try {
                    val linkRef = PointerByReference()
                    COMUtils.checkRC(
                        Ole32.INSTANCE.CoCreateInstance(
                            clsidShellLink,
                            Pointer.NULL,
                            1,
                            iidShellLinkW,
                            linkRef,
                        )
                    )
                    val link = ShellLink(linkRef.value)
                    try {
                        COMUtils.checkRC(link.setPath(executable))
                        COMUtils.checkRC(link.setArguments("--toggle"))
                        COMUtils.checkRC(link.setDescription("Toggle Ad Blocking"))
                        setTaskTitle(link, "Toggle Ad Blocking")
                        COMUtils.checkRC(link.setIconLocation(executable, 0))
                        COMUtils.checkRC(collection.addObject(link.pointer))
                    } finally {
                        link.Release()
                    }
                    COMUtils.checkRC(destination.addUserTasks(collection.pointer))
                    COMUtils.checkRC(destination.commitList())
                } finally {
                    collection.Release()
                }
            } catch (t: Throwable) {
                runCatching { destination.abortList() }
                throw t
            } finally {
                destination.Release()
            }
        } finally {
            if (initialized) Ole32.INSTANCE.CoUninitialize()
        }
    }


    private fun setTaskTitle(link: ShellLink, title: String) {
        val storeRef = PointerByReference()
        COMUtils.checkRC(link.QueryInterface(Guid.REFIID(iidPropertyStore), storeRef))
        val store = PropertyStore(storeRef.value)
        try {
            val key = Memory(20).apply { clear() }
            COMUtils.checkRC(Propsys.INSTANCE.PSPropertyKeyFromString(WString("{F29F85E0-4FF9-1068-AB91-08002B27B3D9} 2"), key))

            val text = Memory(((title.length + 1) * 2).toLong()).apply { setWideString(0, title) }
            val value = Memory(24).apply {
                clear()
                setShort(0, 31)
                setPointer(8, text)
            }
            COMUtils.checkRC(store.setValue(key, value))
            COMUtils.checkRC(store.commit())
        } finally {
            store.Release()
        }
    }
    private fun applicationExecutable(): String {
        val packaged = System.getProperty("jpackage.app-path")?.trim()?.takeIf { it.isNotEmpty() }
        val process = ProcessHandle.current().info().command().orElse(null)?.trim()?.takeIf { it.isNotEmpty() }
        val candidate = packaged ?: process ?: error("Unable to determine BlockAds executable path")
        val absolute = Path.of(candidate).toAbsolutePath().normalize().toString()
        require(absolute.endsWith(".exe", ignoreCase = true)) { "Taskbar shortcut requires packaged BlockAds.exe" }
        return absolute
    }

    private class CustomDestinationList(pointer: Pointer) : Unknown(pointer) {
        fun beginList(slots: IntByReference, iid: Guid.REFIID, removed: PointerByReference): WinNT.HRESULT =
            _invokeNativeObject(4, arrayOf(pointer, slots, iid, removed), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun addUserTasks(tasks: Pointer): WinNT.HRESULT =
            _invokeNativeObject(7, arrayOf(pointer, tasks), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun commitList(): WinNT.HRESULT =
            _invokeNativeObject(8, arrayOf(pointer), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun abortList(): WinNT.HRESULT =
            _invokeNativeObject(11, arrayOf(pointer), WinNT.HRESULT::class.java) as WinNT.HRESULT
    }

    private class ObjectCollection(pointer: Pointer) : Unknown(pointer) {
        fun addObject(value: Pointer): WinNT.HRESULT =
            _invokeNativeObject(5, arrayOf(pointer, value), WinNT.HRESULT::class.java) as WinNT.HRESULT
    }


    private class PropertyStore(pointer: Pointer) : Unknown(pointer) {
        fun setValue(key: Pointer, value: Pointer): WinNT.HRESULT =
            _invokeNativeObject(6, arrayOf(pointer, key, value), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun commit(): WinNT.HRESULT =
            _invokeNativeObject(7, arrayOf(pointer), WinNT.HRESULT::class.java) as WinNT.HRESULT
    }

    private interface Propsys : StdCallLibrary {
        fun PSPropertyKeyFromString(value: WString, key: Pointer): WinNT.HRESULT

        companion object {
            val INSTANCE: Propsys = Native.load("Propsys", Propsys::class.java)
        }
    }
    private class ShellLink(pointer: Pointer) : Unknown(pointer) {
        fun setDescription(value: String): WinNT.HRESULT =
            _invokeNativeObject(7, arrayOf(pointer, WString(value)), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun setArguments(value: String): WinNT.HRESULT =
            _invokeNativeObject(11, arrayOf(pointer, WString(value)), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun setIconLocation(path: String, index: Int): WinNT.HRESULT =
            _invokeNativeObject(17, arrayOf(pointer, WString(path), index), WinNT.HRESULT::class.java) as WinNT.HRESULT

        fun setPath(value: String): WinNT.HRESULT =
            _invokeNativeObject(20, arrayOf(pointer, WString(value)), WinNT.HRESULT::class.java) as WinNT.HRESULT
    }
}
