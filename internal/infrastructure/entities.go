package infrastructure

type TerraformVersion struct {
	TerraformVersion  string `json:"terraform_version"`
	TerraformRevision string `json:"terraform_revision"`
	TerraformOutdated bool   `json:"terraform_outdated"`
}

type TerraformValidateResult struct {
	Valid        bool                   `json:"valid"`
	ErrorCount   int64                  `json:"error_count"`
	WarningCount int64                  `json:"warning_count"`
	Diagnostics  []ValidationDiagnostic `json:"diagnostics"`
}

type TerraformPlan struct {
	FormatVersion    string                  `json:"format_version"`
	TerraformVersion string                  `json:"terraform_version"`
	Variables        map[string]interface{}  `json:"variables"`
	PlannedValues    PlannedValues           `json:"planned_values"`
	ResourceChanges  []ResourceChange        `json:"resource_changes"`
	OutputChanges    map[string]OutputChange `json:"output_changes"`
	PriorState       PriorState              `json:"prior_state"`
}

type OutputChange struct {
	Actions      []Action    `json:"actions"`
	Before       interface{} `json:"before"`
	After        string      `json:"after"`
	AfterUnknown bool        `json:"after_unknown"`
}

type PlannedValues struct {
	Outputs    map[string]Output       `json:"outputs"`
	RootModule PlannedValuesRootModule `json:"root_module"`
}

type Output struct {
	Sensitive bool   `json:"sensitive"`
	Value     string `json:"value"`
}

type PlannedValuesRootModule struct {
	ChildModules []ChildModule `json:"child_modules"`
}

type ChildModule struct {
	Resources []Resource `json:"resources"`
	Address   string     `json:"address"`
}

type Resource struct {
	Address       string     `json:"address"`
	Mode          string     `json:"mode"`
	Type          string     `json:"type"`
	Name          string     `json:"name"`
	ProviderName  string     `json:"provider_name"`
	SchemaVersion int64      `json:"schema_version"`
	Values        AfterClass `json:"values"`
}

type AfterClass struct {
	Location   string      `json:"location,omitempty"`
	Name       string      `json:"name,omitempty"`
	Tags       interface{} `json:"tags"`
	Timeouts   interface{} `json:"timeouts"`
	Keepers    interface{} `json:"keepers"`
	Prefix     interface{} `json:"prefix"`
	ByteLength *int64      `json:"byte_length,omitempty"`
}

type PriorState struct {
	FormatVersion    string `json:"format_version"`
	TerraformVersion string `json:"terraform_version"`
	Values           Values `json:"values"`
}

type Values struct {
	Outputs    map[string]interface{} `json:"outputs"`
	RootModule ValueClass             `json:"root_module"`
}

type ValueClass struct {
}

type ResourceChange struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	ProviderName  string `json:"provider_name"`
	Change        Change `json:"change"`
}

type Change struct {
	Actions      []Action     `json:"actions"`
	Before       interface{}  `json:"before"`
	After        AfterClass   `json:"after"`
	AfterUnknown AfterUnknown `json:"after_unknown"`
}

type AfterUnknown struct {
	ID     bool  `json:"id"`
	B64Std *bool `json:"b64_std,omitempty"`
	B64URL *bool `json:"b64_url,omitempty"`
	DEC    *bool `json:"dec,omitempty"`
	Hex    *bool `json:"hex,omitempty"`
}

type AadAdminGroups struct {
	Value []interface{} `json:"value"`
}

type GlobalRglocation struct {
	Value string `json:"value"`
}

type AgentManagementRgEnabled struct {
	Value bool `json:"value"`
}

type Tags struct {
	Value ValueClass `json:"value"`
}

type Action string

const (
	Create Action = "create"
	Delete Action = "delete"
	NoOp   Action = "no-op"
	Update Action = "update"
)

type TerraformOutputVariable struct {
	Sensitive bool   `json:"sensitive"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

type ValidationDiagnostic struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
	Range    Range  `json:"range"`
}

type Range struct {
	Filename string `json:"filename"`
	Start    End    `json:"start"`
	End      End    `json:"end"`
}

type End struct {
	Line   int64 `json:"line"`
	Column int64 `json:"column"`
	Byte   int64 `json:"byte"`
}

type PlanChanges struct {
	CreateOps int
	ChangeOps int
	DeleteOps int
	NoOps     int
}

func (p PlanChanges) HasChanges() bool {
	if p.ChangeOps > 0 || p.CreateOps > 0 || p.DeleteOps > 0 {
		return true
	}

	return false
}
func (p PlanChanges) HasDestructiveChanges() bool {
	return p.DeleteOps > 0
}
