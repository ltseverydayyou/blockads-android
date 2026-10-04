package app.pwhs.blockads.ui.filter.component

import androidx.compose.animation.animateContentSize
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import app.pwhs.blockads.data.entities.FilterList

@Composable
fun FilterListCard(
    filters: List<FilterList>,
    onToggle: (FilterList) -> Unit,
    onFilterClick: (Long) -> Unit,
    onDelete: ((FilterList) -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    if (filters.isEmpty()) return

    Card(
        modifier = modifier
            .fillMaxWidth()
            .animateContentSize(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column {
            filters.forEachIndexed { index, filter ->
                FilterItem(
                    filter = filter,
                    onToggle = { onToggle(filter) },
                    onDelete = onDelete?.let { { it(filter) } },
                    onClick = { onFilterClick(filter.id) }
                )
                if (index < filters.lastIndex) {
                    HorizontalDivider(
                        modifier = Modifier.padding(horizontal = 16.dp),
                        color = MaterialTheme.colorScheme.outline.copy(alpha = 0.1f)
                    )
                }
            }
        }
    }
}
