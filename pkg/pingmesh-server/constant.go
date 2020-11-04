package pingmesh_server

const (
	NodeRoleKey                = "node-role.kubernetes.io/patition"
	NodeRoleValue              = ""
	DefaultPingmeshDownloadURL = "/getPinglist"
	DefaultPingmeshUploadURL   = "/uploadMetrics"
	DefaultMetricsURL          = "/metrics"
	DefaultHttpsAddress        = "0.0.0.0"
	DefaultHttpsPort           = 10009
	DefaultMetricsPort         = 6002

	// ping
	MetricsNamePingLatency       = `ping_latency_millonseconds`
	MetricsNamePingPackageDrop   = `ping_packageDrop_rate`
	MetricsNamePingTargetSuccess = `ping_target_success`
)
