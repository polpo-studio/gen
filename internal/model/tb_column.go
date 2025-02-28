package model

import (
	"bytes"
	"fmt"
	"github.com/polpo-studio/gen/internal/utils"
	"strings"
)

// Column table column's info
type Column struct {
	TableName     string   `gorm:"column:TABLE_NAME"`
	ColumnName    string   `gorm:"column:COLUMN_NAME"`
	ColumnComment string   `gorm:"column:COLUMN_COMMENT"`
	DataType      string   `gorm:"column:DATA_TYPE"`
	ColumnKey     string   `gorm:"column:COLUMN_KEY"`
	ColumnType    string   `gorm:"column:COLUMN_TYPE"`
	ColumnDefault string   `gorm:"column:COLUMN_DEFAULT"`
	Extra         string   `gorm:"column:EXTRA"`
	IsNullable    string   `gorm:"column:IS_NULLABLE"`
	Indexes       []*Index `gorm:"-"`

	dataTypeMap map[string]func(detailType string) (dataType string) `gorm:"-"`
}

func (c *Column) IsPrimaryKey() bool {
	return c != nil && c.ColumnKey == "PRI"
}

func (c *Column) AutoIncrement() bool {
	return c != nil && c.Extra == "auto_increment"
}

func (c *Column) SetDataTypeMap(m map[string]func(detailType string) (dataType string)) {
	c.dataTypeMap = m
}

func (c *Column) GetDataType() (memberType string) {
	if mapping, ok := c.dataTypeMap[c.DataType]; ok {
		return mapping(c.ColumnType)
	}
	return dataType.Get(c.DataType, c.ColumnType)
}

func (c *Column) ToMember(nullable bool) *Member {
	memberType := c.GetDataType()
	if c.ColumnName == "deleted_at" && memberType == "time.Time" {
		memberType = "gorm.DeletedAt"
	} else if nullable && c.IsNullable == "YES" {
		memberType = "*" + memberType
	}
	return &Member{
		Name:             c.ColumnName,
		Type:             memberType,
		ColumnName:       c.ColumnName,
		ColumnComment:    c.ColumnComment,
		MultilineComment: c.multilineComment(),
		GORMTag:          c.buildGormTag(),
		JSONTag:          c.buildJsonTag(),
	}
}

func (c *Column) buildJsonTag() string {
	// converts a string to lowerCamelCase
	return utils.ConvertCamelInitCase(c.ColumnName, false)
}

func (c *Column) multilineComment() bool { return strings.Contains(c.ColumnComment, "\n") }

func (c *Column) buildGormTag() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("column:%s;type:%s", c.ColumnName, c.ColumnType))
	if c.IsPrimaryKey() {
		buf.WriteString(";primaryKey")
		if !c.AutoIncrement() {
			// integer PrioritizedPrimaryField enables AutoIncrement by default,
			// if not, we need to turn off autoIncrement for the fields
			buf.WriteString(";autoIncrement:false")
		}
	} else if c.IsNullable != "YES" {
		buf.WriteString(";not null")
	}

	for _, idx := range c.Indexes {
		if idx == nil || idx.IsPrimaryKey() {
			continue
		}
		if idx.IsUnique() {
			buf.WriteString(fmt.Sprintf(";uniqueIndex:%s,priority:%d", idx.IndexName, idx.SeqInIndex))
		} else {
			buf.WriteString(fmt.Sprintf(";index:%s,priority:%d", idx.IndexName, idx.SeqInIndex))
		}
	}
	if c.ColumnDefault != "" {
		buf.WriteString(fmt.Sprintf(";default:%s", c.ColumnDefault))
	}
	return buf.String()
}
