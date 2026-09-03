import org.jetbrains.compose.desktop.application.dsl.TargetFormat

plugins {
    kotlin("jvm") version "2.3.10"
    kotlin("plugin.serialization") version "2.3.10"
    id("org.jetbrains.compose") version "1.10.2"
    id("org.jetbrains.kotlin.plugin.compose") version "2.3.10"
}

group = "app.pwhs.blockads"
version = "1.2.0"

kotlin {
    jvmToolchain(21)
}

dependencies {
    implementation(compose.desktop.currentOs)
    implementation(compose.material3)
    implementation(compose.materialIconsExtended)
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.11.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.10.2")
}

compose.desktop {
    application {
        mainClass = "app.pwhs.blockads.desktop.MainKt"
        nativeDistributions {
            modules("java.net.http")
            targetFormats(TargetFormat.Exe)
            packageName = "BlockAds"
            packageVersion = "1.2.0"
            description = "BlockAds for Windows"
            vendor = "BlockAds"
            windows {
                iconFile.set(project.file("blockads.ico"))
                menuGroup = "BlockAds"
                shortcut = true
                dirChooser = true
                perUserInstall = true
            }
        }
    }
}
