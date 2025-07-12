package types

import (
	"time"
)

type CapsuleBlueprint struct {
	BaseModel
	ParentID         string           `json:"parent_id" gorm:"type:varchar(255);column:parent_id"`
	Name             string           `json:"name" gorm:"not null;type:varchar(255);column:name"`
	Slug             string           `json:"slug" gorm:"not null;type:varchar(255);column:slug"`
	Type             string           `json:"type" gorm:"not null;type:varchar(255);column:type"`
	Version          string           `json:"version" gorm:"not null;type:varchar(255);column:version"`
	Active           bool             `json:"active" gorm:"not null;type:boolean;column:active"`
	Services         []CapsuleService `json:"services" gorm:"foreignKey:CapsuleID;constraint:OnDelete:CASCADE"`
	Files            []CapsuleFile    `json:"files" gorm:"foreignKey:CapsuleID;constraint:OnDelete:CASCADE"`
	SetupScript      []byte           `json:"setup_script" gorm:"column:setup_script"`
	LastDeployedAt   time.Time        `json:"last_deployed_at" gorm:"type:timestamp;column:last_deployed_at"`
	LastDownloadedAt time.Time        `json:"last_downloaded_at" gorm:"type:timestamp;column:last_downloaded_at"`
	DownloadCount    int              `json:"download_count" gorm:"not null;type:integer;column:download_count;default:0"`
}

func (c *CapsuleBlueprint) TableName() string {
	return "capsule_blueprints"
}

type CapsuleFile struct {
	BaseModel
	CapsuleID string `json:"capsule_id" gorm:"not null;type:varchar(255);column:capsule_id;index"`
	FileName  string `json:"file_name" gorm:"not null;type:varchar(255);column:file_name"`
	Path      string `json:"path" gorm:"not null;type:varchar(255);column:path"`
	Content   []byte `json:"content" gorm:"not null;type:json;column:content"`
	UID       int    `json:"uid" gorm:"not null;type:integer;column:uid"`
	GID       int    `json:"gid" gorm:"not null;type:integer;column:gid"`
	Mode      int    `json:"mode" gorm:"not null;type:integer;column:mode"`
}

func (c *CapsuleFile) TableName() string {
	return "capsule_files"
}

type CapsuleService struct {
	BaseModel
	CapsuleID       string `json:"blueprint_id" gorm:"not null;type:varchar(255);column:blueprint_id;index"`
	ServiceName     string `json:"service_name" gorm:"not null;type:varchar(255);column:service_name"`
	Slug            string `json:"slug" gorm:"not null;type:varchar(255);column:slug"`
	ContainerConfig string `json:"container_config" gorm:"not null;type:json;column:container_config"`
	Parameters      string `json:"parameters" gorm:"not null;type:json;column:parameters"`
	Volumes         string `json:"volumes" gorm:"not null;type:json;column:volumes"`
	PortMappings    string `json:"port_mappings" gorm:"not null;type:json;column:port_mappings"`
}

func (c *CapsuleService) TableName() string {
	return "capsule_services"
}
