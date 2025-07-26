package models

type Context struct {
	BaseModel
	Name              string                 `json:"name" gorm:"column:name,type:varchar(255),not null,unique"`
	Description       string                 `json:"description" gorm:"column:description,type:text"`
	IsDefault         bool                   `json:"is_default" gorm:"column:is_default,type:boolean,default:false"`
	IsEnabled         bool                   `json:"is_enabled" gorm:"column:is_enabled,type:boolean,default:true"`
	ConfigPath        string                 `json:"config_path" gorm:"column:config_path,type:varchar(255),not null"`
	MonitorForChanges bool                   `json:"monitor_for_changes" gorm:"column:monitor_for_changes,type:boolean,default:false"`
	Config            ContextConfig          `json:"config" gorm:"column:config,type:json"`
	Services          map[string]interface{} `json:"services" gorm:"column:services,type:json"`
}

type ContextConfig struct {
	BaseModel
	ContextID    string                   `json:"context_id" gorm:"column:context_id,type:uuid,not null"`
	Network      ContextNetwork           `json:"network" gorm:"column:network,type:json"`
	Cors         ContextCorsConfiguration `json:"cors" gorm:"column:cors,type:json"`
	Certificates []RootCertificate        `json:"certificates" gorm:"column:certificates,type:json"`
}

type ContextCorsConfiguration struct {
	AllowedMethods   []string `json:"allowedMethods" yaml:"allowedMethods" gorm:"column:allowed_methods,type:json,not null"`
	AllowedHeaders   []string `json:"allowedHeaders" yaml:"allowedHeaders" gorm:"column:allowed_headers,type:json,not null"`
	AllowedOrigins   []string `json:"allowedOrigins" yaml:"allowedOrigins" gorm:"column:allowed_origins,type:json,not null"`
	AllowCredentials bool     `json:"allowCredentials" yaml:"allowCredentials" gorm:"column:allow_credentials,type:boolean,default:false"`
	MaxAge           int      `json:"maxAge" yaml:"maxAge" gorm:"column:max_age,type:int,default:0"`
}

type ContextNetwork struct {
	LocalIP     string      `json:"localIp,omitempty" yaml:"localIp,omitempty" gorm:"column:local_ip,type:varchar(255),not null"`
	DomainName  string      `json:"domainName,omitempty" yaml:"domainName,omitempty" gorm:"column:domain_name,type:varchar(255),not null"`
	Certificate Certificate `json:"certificatePath,omitempty" yaml:"certificatePath,omitempty" gorm:"column:certificate,type:json,not null"`
}
