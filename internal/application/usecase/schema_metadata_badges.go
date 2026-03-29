package usecase

import "github.com/mgierok/dbc/internal/domain/model"

func projectedSchemaMetadataBadges(column model.Column) []string {
	badges := make([]string, 0, 4+len(column.ForeignKeys))

	if column.PrimaryKey {
		badges = append(badges, "PK")
	}

	if column.Nullable {
		badges = append(badges, "NULL")
	} else {
		badges = append(badges, "NOT NULL")
	}

	if column.Unique && !column.PrimaryKey {
		badges = append(badges, "UNIQUE")
	}

	if column.DefaultValue != nil {
		badges = append(badges, "DEFAULT "+*column.DefaultValue)
	}

	if column.AutoIncrement {
		badges = append(badges, "AUTOINCREMENT")
	}

	for _, foreignKey := range column.ForeignKeys {
		badges = append(badges, projectedForeignKeyBadge(foreignKey))
	}

	return badges
}

func projectedForeignKeyBadge(foreignKey model.ForeignKeyRef) string {
	if foreignKey.Column == "" {
		return "FK->" + foreignKey.Table
	}

	return "FK->" + foreignKey.Table + "." + foreignKey.Column
}
