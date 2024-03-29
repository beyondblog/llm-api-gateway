package provider

import (
	"encoding/json"
	"fmt"
	"github.com/beyondblog/llm-api-gateway/utils"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type VastAIProvider struct {
	apiKey string
	model  string
	branch string
	label  string
}

type ExecuteCommandResponse struct {
	Success       bool   `json:"success"`
	WriteablePath string `json:"writeable_path"`
	ResultUrl     string `json:"result_url"`
	Msg           string `json:"msg"`
}

type InstanceResponse struct {
	Instances []Instance `json:"instances"`
}

type LLMModelInfoResponse struct {
	ModelName string        `json:"model_name"`
	LoraNames []interface{} `json:"lora_names"`
}

type Instance struct {
	IsBid             bool                          `json:"is_bid"`
	InetUpBilled      float64                       `json:"inet_up_billed"`
	InetDownBilled    float64                       `json:"inet_down_billed"`
	External          bool                          `json:"external"`
	Webpage           interface{}                   `json:"webpage"`
	Rentable          bool                          `json:"rentable"`
	ComputeCap        int                           `json:"compute_cap"`
	CreditBalance     interface{}                   `json:"credit_balance"`
	CreditDiscount    interface{}                   `json:"credit_discount"`
	CreditDiscountMax float64                       `json:"credit_discount_max"`
	DriverVersion     string                        `json:"driver_version"`
	CudaMaxGood       float64                       `json:"cuda_max_good"`
	MachineId         int                           `json:"machine_id"`
	HostingType       int                           `json:"hosting_type"`
	PublicIpaddr      string                        `json:"public_ipaddr"`
	Geolocation       string                        `json:"geolocation"`
	FlopsPerDphtotal  float64                       `json:"flops_per_dphtotal"`
	DlperfPerDphtotal float64                       `json:"dlperf_per_dphtotal"`
	Reliability2      float64                       `json:"reliability2"`
	HostRunTime       float64                       `json:"host_run_time"`
	ClientRunTime     float64                       `json:"client_run_time"`
	HostId            int                           `json:"host_id"`
	Id                int                           `json:"id"`
	BundleId          int                           `json:"bundle_id"`
	NumGpus           int                           `json:"num_gpus"`
	TotalFlops        float64                       `json:"total_flops"`
	MinBid            float64                       `json:"min_bid"`
	DphBase           float64                       `json:"dph_base"`
	DphTotal          float64                       `json:"dph_total"`
	GpuName           string                        `json:"gpu_name"`
	GpuRam            int                           `json:"gpu_ram"`
	GpuTotalram       int                           `json:"gpu_totalram"`
	VramCostperhour   float64                       `json:"vram_costperhour"`
	GpuDisplayActive  bool                          `json:"gpu_display_active"`
	GpuMemBw          float64                       `json:"gpu_mem_bw"`
	BwNvlink          float64                       `json:"bw_nvlink"`
	DirectPortCount   int                           `json:"direct_port_count"`
	GpuLanes          int                           `json:"gpu_lanes"`
	PcieBw            float64                       `json:"pcie_bw"`
	PciGen            float64                       `json:"pci_gen"`
	Dlperf            float64                       `json:"dlperf"`
	CpuName           string                        `json:"cpu_name"`
	MoboName          string                        `json:"mobo_name"`
	CpuRam            int                           `json:"cpu_ram"`
	CpuCores          int                           `json:"cpu_cores"`
	CpuCoresEffective float64                       `json:"cpu_cores_effective"`
	GpuFrac           float64                       `json:"gpu_frac"`
	HasAvx            int                           `json:"has_avx"`
	DiskSpace         float64                       `json:"disk_space"`
	DiskName          string                        `json:"disk_name"`
	DiskBw            float64                       `json:"disk_bw"`
	InetUp            float64                       `json:"inet_up"`
	InetDown          float64                       `json:"inet_down"`
	StartDate         float64                       `json:"start_date"`
	EndDate           *float64                      `json:"end_date"`
	Duration          *float64                      `json:"duration"`
	StorageCost       float64                       `json:"storage_cost"`
	InetUpCost        float64                       `json:"inet_up_cost"`
	InetDownCost      float64                       `json:"inet_down_cost"`
	StorageTotalCost  float64                       `json:"storage_total_cost"`
	OsVersion         string                        `json:"os_version"`
	Verification      string                        `json:"verification"`
	StaticIp          bool                          `json:"static_ip"`
	Score             float64                       `json:"score"`
	SshIdx            string                        `json:"ssh_idx"`
	SshHost           string                        `json:"ssh_host"`
	SshPort           int                           `json:"ssh_port"`
	ActualStatus      string                        `json:"actual_status"`
	IntendedStatus    string                        `json:"intended_status"`
	CurState          string                        `json:"cur_state"`
	NextState         string                        `json:"next_state"`
	ImageUuid         string                        `json:"image_uuid"`
	ImageArgs         []interface{}                 `json:"image_args"`
	ImageRuntype      string                        `json:"image_runtype"`
	ExtraEnv          [][]string                    `json:"extra_env"`
	Onstart           string                        `json:"onstart"`
	Label             string                        `json:"label"`
	JupyterToken      string                        `json:"jupyter_token"`
	StatusMsg         string                        `json:"status_msg"`
	GpuUtil           float64                       `json:"gpu_util"`
	DiskUtil          float64                       `json:"disk_util"`
	DiskUsage         float64                       `json:"disk_usage"`
	GpuTemp           float64                       `json:"gpu_temp"`
	LocalIpaddrs      string                        `json:"local_ipaddrs"`
	DirectPortEnd     int                           `json:"direct_port_end"`
	DirectPortStart   int                           `json:"direct_port_start"`
	CpuUtil           float64                       `json:"cpu_util"`
	MemUsage          float64                       `json:"mem_usage"`
	MemLimit          float64                       `json:"mem_limit"`
	VmemUsage         float64                       `json:"vmem_usage"`
	MachineDirSshPort int                           `json:"machine_dir_ssh_port"`
	Ports             map[string][]InstancePortInfo `json:"ports"`
}

type InstancePortInfo struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type ExecuteCommand struct {
	Command string `json:"command"`
}

type QueryBundleResponse struct {
	Offers []struct {
		IsBid              bool        `json:"is_bid"`
		InetUpBilled       interface{} `json:"inet_up_billed"`
		InetDownBilled     interface{} `json:"inet_down_billed"`
		External           bool        `json:"external"`
		Webpage            interface{} `json:"webpage"`
		Logo               string      `json:"logo"`
		Rentable           bool        `json:"rentable"`
		ComputeCap         int         `json:"compute_cap"`
		CreditBalance      interface{} `json:"credit_balance"`
		CreditDiscount     interface{} `json:"credit_discount"`
		CreditDiscountMax  *float64    `json:"credit_discount_max"`
		DriverVersion      string      `json:"driver_version"`
		CudaMaxGood        float64     `json:"cuda_max_good"`
		MachineId          int         `json:"machine_id"`
		HostingType        *int        `json:"hosting_type"`
		PublicIpaddr       string      `json:"public_ipaddr"`
		Geolocation        string      `json:"geolocation"`
		FlopsPerDphtotal   float64     `json:"flops_per_dphtotal"`
		DlperfPerDphtotal  float64     `json:"dlperf_per_dphtotal"`
		Reliability2       float64     `json:"reliability2"`
		HostRunTime        float64     `json:"host_run_time"`
		ClientRunTime      float64     `json:"client_run_time"`
		HostId             int         `json:"host_id"`
		Id                 int         `json:"id"`
		BundleId           int         `json:"bundle_id"`
		NumGpus            int         `json:"num_gpus"`
		TotalFlops         float64     `json:"total_flops"`
		MinBid             float64     `json:"min_bid"`
		DphBase            float64     `json:"dph_base"`
		DphTotal           float64     `json:"dph_total"`
		GpuName            string      `json:"gpu_name"`
		GpuRam             int         `json:"gpu_ram"`
		GpuTotalram        int         `json:"gpu_totalram"`
		VramCostperhour    float64     `json:"vram_costperhour"`
		GpuDisplayActive   bool        `json:"gpu_display_active"`
		GpuMemBw           float64     `json:"gpu_mem_bw"`
		BwNvlink           float64     `json:"bw_nvlink"`
		DirectPortCount    int         `json:"direct_port_count"`
		GpuLanes           int         `json:"gpu_lanes"`
		PcieBw             float64     `json:"pcie_bw"`
		PciGen             float64     `json:"pci_gen"`
		Dlperf             float64     `json:"dlperf"`
		CpuName            string      `json:"cpu_name"`
		MoboName           string      `json:"mobo_name"`
		CpuRam             int         `json:"cpu_ram"`
		CpuCores           int         `json:"cpu_cores"`
		CpuCoresEffective  float64     `json:"cpu_cores_effective"`
		GpuFrac            float64     `json:"gpu_frac"`
		HasAvx             int         `json:"has_avx"`
		DiskSpace          float64     `json:"disk_space"`
		DiskName           string      `json:"disk_name"`
		DiskBw             float64     `json:"disk_bw"`
		InetUp             float64     `json:"inet_up"`
		InetDown           float64     `json:"inet_down"`
		StartDate          float64     `json:"start_date"`
		EndDate            *float64    `json:"end_date"`
		Duration           *float64    `json:"duration"`
		StorageCost        float64     `json:"storage_cost"`
		InetUpCost         float64     `json:"inet_up_cost"`
		InetDownCost       float64     `json:"inet_down_cost"`
		StorageTotalCost   float64     `json:"storage_total_cost"`
		OsVersion          string      `json:"os_version"`
		Verification       string      `json:"verification"`
		StaticIp           bool        `json:"static_ip"`
		Score              float64     `json:"score"`
		DiscountRate       *float64    `json:"discount_rate"`
		DiscountedHourly   float64     `json:"discounted_hourly"`
		DiscountedDphTotal float64     `json:"discounted_dph_total"`
		Rented             bool        `json:"rented"`
		BundledResults     int         `json:"bundled_results"`
		PendingCount       int         `json:"pending_count"`
	} `json:"offers"`
}

type CreateInstanceParam struct {
	ClientId      string            `json:"client_id"`
	Image         string            `json:"image"`
	Env           map[string]string `json:"env"`
	Price         float64           `json:"price"`
	Disk          float64           `json:"disk"`
	Label         string            `json:"label"`
	Onstart       string            `json:"onstart"`
	RunType       string            `json:"runtype"`
	ImageLogin    string            `json:"image_login;omitempty"`
	PythonUtf8    bool              `json:"python_utf8"`
	LangUtf8      bool              `json:"lang_utf8"`
	UseJupyterLab bool              `json:"use_jupyter_lab"`
	JupyterDir    string            `json:"jupyter_dir"`
	CreateFrom    string            `json:"create_from"`
	Force         bool              `json:"force"`
}

func NewVastAIProvider(apiKey, model, branch, label string) *VastAIProvider {
	if branch == "" {
		branch = "main"
	}
	return &VastAIProvider{
		apiKey: apiKey,
		branch: branch,
		model:  model,
		label:  label,
	}
}

func (v *VastAIProvider) GetModel() string {
	return fmt.Sprintf("%s_%s", strings.ReplaceAll(v.model, "/", "_"), v.branch)
}

func (v *VastAIProvider) GetEndpoints() ([]ServerEndpoint, error) {
	data, err := v.request("GET", "https://console.vast.ai/api/v0/instances", nil)
	if err != nil {
		return nil, err
	}
	var instanceResponse InstanceResponse
	err = json.Unmarshal(data, &instanceResponse)
	if err != nil {
		return nil, err
	}

	var endpoints []ServerEndpoint
	for _, instance := range instanceResponse.Instances {
		if !strings.Contains(instance.ImageUuid, "text-generation-webui") {
			continue
		}
		// text-generation-webui check if port 5000 is open
		if instance.Ports["5000/tcp"] == nil {
			continue
		}
		port, _ := strconv.Atoi(instance.Ports["5000/tcp"][0].HostPort)

		if instance.Label != v.label {
			continue
		}

		endpoint := ServerEndpoint{
			ID:      fmt.Sprintf("%d", instance.Id),
			Host:    strings.TrimSpace(instance.PublicIpaddr),
			Port:    port,
			CPUName: instance.CpuName,
			GPUName: instance.GpuName,
		}

		if v.healthCheck(endpoint) {
			endpoints = append(endpoints, endpoint)
		}

	}
	return endpoints, nil
}

func (v *VastAIProvider) AutoScaling(replica int) error {
	//TODO implement me
	panic("implement me")
}

func (v *VastAIProvider) request(method, url string, payload []byte) ([]byte, error) {
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", v.apiKey),
	}
	if method == "GET" {
		return utils.GetHttpRequest(url, header)
	} else if method == "POST" {
		return utils.PostHttpRequest(url, header, payload)
	} else if method == "PUT" {
		return utils.PutHttpRequest(url, header, payload)
	}
	return nil, fmt.Errorf("unsupported method: %s", method)
}

func (v *VastAIProvider) executeCommand(instanceID int, command string) (*ExecuteCommandResponse, error) {

	commandParam := ExecuteCommand{
		Command: command,
	}
	payload, _ := json.Marshal(commandParam)

	data, err := v.request("PUT", fmt.Sprintf("https://console.vast.ai/api/v0/instances/command/%d/", instanceID), payload)
	if err != nil {
		return nil, err
	}

	var response ExecuteCommandResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (v *VastAIProvider) healthCheck(endpoint ServerEndpoint) bool {
	data, err := v.request("GET", fmt.Sprintf("http://%s:%d/v1/internal/model/info", endpoint.Host,
		endpoint.Port), nil)
	if err != nil {
		log.Printf("endpoint health check err: %v\n, %s:%d", err, endpoint.Host, endpoint.Port)
		return false
	}

	var response LLMModelInfoResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Printf("endpoint health check err: %v\n, %s:%d", err, endpoint.Host, endpoint.Port)
		return false
	}
	return response.ModelName == v.GetModel()
}

func (v *VastAIProvider) createInstance() error {

	diskSpace := "30"
	gpuName := "RTX 4090"
	dlperf := "70"
	query := fmt.Sprintf(`{"verified": {"eq": "True"}, "external": {"eq": false}, "rentable": {"eq": true}, "rented": {"eq": false}, "disk_space": {"gte": "%s"}, "gpu_name": {"eq": "%s"}, "num_gpus": {"eq": "1"}, "reliability2": {"gt": "0.99"}, "dlperf": {"gt": "%s"}, "order": [["dphtotal", "asc"], ["total_flops", "asc"]], "type": "on-demand"}`,
		diskSpace, gpuName, dlperf)
	query = url.QueryEscape(query)
	data, err := v.request("GET", fmt.Sprintf("https://console.vast.ai/api/v0/bundles?q=%s", query), nil)
	if err != nil {
		return err
	}

	var response QueryBundleResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	log.Printf("%+v", response.Offers[0])

	createParam := CreateInstanceParam{
		ClientId:      "me",
		Image:         "atinoda/text-generation-webui:default-snapshot-2023-12-31",
		Env:           nil,
		Price:         response.Offers[0].DphTotal,
		Disk:          30.48,
		Label:         "",
		Onstart:       "env | grep _ >> /etc/environment; pip install accelerate -U; pip install protobuf;python3 /app/download-model.py --output /app/models --branch gptq-4bit-32g-actorder_True TheBloke/U-Amethyst-20B-GPTQ; cd /app; /scripts/docker-entrypoint.sh python3 /app/server.py --listen --api --verbose --loader ExLlamav2_HF --model TheBloke_U-Amethyst-20B-GPTQ_gptq-4bit-32g-actorder_True;",
		RunType:       "jupyter_direc ssh_direc ssh_proxy",
		ImageLogin:    "",
		PythonUtf8:    false,
		LangUtf8:      false,
		UseJupyterLab: false,
		JupyterDir:    "",
		CreateFrom:    "",
		Force:         false,
	}
	machineID := response.Offers[0].Id
	payload, _ := json.Marshal(createParam)

	log.Printf(string(payload))

	data, err = v.request("PUT", fmt.Sprintf("https://console.vast.ai/api/v0/asks/%d/", machineID), payload)
	if err != nil {
		return err
	}

	log.Println(string(data))
	return nil
}
