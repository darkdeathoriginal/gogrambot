package models

type ExternalPlugin struct {
	Name string
	Url  string `gorm:"uniqueIndex"`
}

func (ExternalPlugin) TableName() string {
	return "external_plugin"
}

func init() {
	AddModelToMigrate(&ExternalPlugin{})
}
