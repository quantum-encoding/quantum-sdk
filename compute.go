package qai

import (
	"context"
	"fmt"
)

// ComputeTemplate describes an available machine configuration with pricing.
type ComputeTemplate struct {
	// ID is the template identifier (e.g. "gpu-t4-standard").
	ID string `json:"id"`

	// DisplayName is a human-readable name.
	DisplayName string `json:"display_name"`

	// MachineType is the GCE machine type (e.g. "n1-standard-8").
	MachineType string `json:"machine_type"`

	// GPUType is the GPU accelerator type (e.g. "nvidia-tesla-t4").
	GPUType string `json:"gpu_type,omitempty"`

	// GPUCount is the number of GPUs attached.
	GPUCount int `json:"gpu_count,omitempty"`

	// VCPUs is the number of virtual CPUs.
	VCPUs int `json:"vcpus"`

	// MemoryGB is the amount of RAM in gigabytes.
	MemoryGB int `json:"memory_gb"`

	// DiskGB is the boot disk size in gigabytes.
	DiskGB int `json:"disk_gb"`

	// HourlyUSD is the on-demand hourly rate in US dollars.
	HourlyUSD float64 `json:"hourly_usd"`

	// SpotHourlyUSD is the spot/preemptible hourly rate (if available).
	SpotHourlyUSD float64 `json:"spot_hourly_usd,omitempty"`

	// SpotAllowed indicates whether spot instances are available.
	SpotAllowed bool `json:"spot_allowed,omitempty"`

	// AvailableZones lists the GCP zones where this template can be provisioned.
	AvailableZones []string `json:"available_zones,omitempty"`

	// BootTimeSecs is the estimated boot time in seconds.
	BootTimeSecs int `json:"boot_time_secs,omitempty"`
}

// TemplatesResponse is the response from listing compute templates.
type TemplatesResponse struct {
	// Templates is the list of available machine configurations.
	Templates []ComputeTemplate `json:"templates"`
}

// ProvisionRequest is the request body for provisioning a compute instance.
type ProvisionRequest struct {
	// Template is the template ID (required).
	Template string `json:"template"`

	// Zone is the GCP zone (optional — defaults to first available zone).
	Zone string `json:"zone,omitempty"`

	// Spot requests a preemptible/spot instance for lower cost.
	Spot bool `json:"spot,omitempty"`

	// AutoTeardownMinutes is the inactivity timeout before auto-teardown (default 30).
	AutoTeardownMinutes int `json:"auto_teardown_minutes,omitempty"`

	// SSHPublicKey is an optional SSH public key to inject at boot.
	SSHPublicKey string `json:"ssh_public_key,omitempty"`
}

// ProvisionResponse is the response from provisioning a compute instance.
type ProvisionResponse struct {
	// InstanceID is the unique instance identifier for subsequent API calls.
	InstanceID string `json:"instance_id"`

	// Status is the initial instance status (e.g. "provisioning").
	Status string `json:"status"`

	// Zone is the GCP zone where the instance is being created.
	Zone string `json:"zone"`

	// MachineType is the GCE machine type.
	MachineType string `json:"machine_type"`

	// GPUType is the GPU accelerator type.
	GPUType string `json:"gpu_type,omitempty"`

	// HourlyUSD is the hourly rate being charged.
	HourlyUSD float64 `json:"hourly_usd"`

	// CostUSD is the initial cost (1-hour minimum).
	CostUSD float64 `json:"cost_usd"`

	// EstimatedBootSecs is the estimated time to boot.
	EstimatedBootSecs int `json:"estimated_boot_secs,omitempty"`
}

// ComputeInstanceInfo describes a compute instance.
type ComputeInstanceInfo struct {
	// InstanceID is the unique instance identifier.
	InstanceID string `json:"instance_id"`

	// Template is the template that was used.
	Template string `json:"template"`

	// Status is the current instance status ("provisioning", "running", "stopping", "terminated", "failed").
	Status string `json:"status"`

	// GCPStatus is the live GCE instance status (populated for running instances).
	GCPStatus string `json:"gcp_status,omitempty"`

	// Zone is the GCP zone.
	Zone string `json:"zone"`

	// MachineType is the GCE machine type.
	MachineType string `json:"machine_type,omitempty"`

	// ExternalIP is the public IP address (available once running).
	ExternalIP string `json:"external_ip,omitempty"`

	// GPUType is the GPU accelerator type.
	GPUType string `json:"gpu_type,omitempty"`

	// GPUCount is the number of GPUs.
	GPUCount int `json:"gpu_count,omitempty"`

	// Spot indicates whether this is a spot/preemptible instance.
	Spot bool `json:"spot,omitempty"`

	// HourlyUSD is the hourly rate.
	HourlyUSD float64 `json:"hourly_usd"`

	// CostUSD is the total cost so far.
	CostUSD float64 `json:"cost_usd"`

	// UptimeMinutes is the total uptime in minutes.
	UptimeMinutes int `json:"uptime_minutes"`

	// AutoTeardownMinutes is the inactivity timeout.
	AutoTeardownMinutes int `json:"auto_teardown_minutes"`

	// SSHUsername is the SSH username for the instance.
	SSHUsername string `json:"ssh_username,omitempty"`

	// LastActiveAt is the ISO 8601 timestamp of last activity.
	LastActiveAt string `json:"last_active_at,omitempty"`

	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// TerminatedAt is the ISO 8601 termination timestamp (if terminated).
	TerminatedAt string `json:"terminated_at,omitempty"`

	// ErrorMessage contains the error if the instance failed.
	ErrorMessage string `json:"error_message,omitempty"`
}

// InstancesResponse is the response from listing compute instances.
type InstancesResponse struct {
	// Instances is the list of compute instances.
	Instances []ComputeInstanceInfo `json:"instances"`
}

// InstanceResponse is the response from getting a single compute instance.
type InstanceResponse = ComputeInstanceInfo

// DeleteResponse is the response from deleting a compute instance.
type DeleteResponse struct {
	// InstanceID is the instance that was terminated.
	InstanceID string `json:"instance_id"`

	// Status is "terminated".
	Status string `json:"status"`

	// FinalCostUSD is the total cost for the instance's lifetime.
	FinalCostUSD float64 `json:"final_cost_usd"`

	// UptimeMinutes is the total uptime.
	UptimeMinutes int `json:"uptime_minutes"`
}

// SSHKeyRequest is the request body for injecting an SSH key into a running instance.
type SSHKeyRequest struct {
	// PublicKey is the SSH public key to inject (required).
	PublicKey string `json:"public_key"`

	// Username is the SSH username (default "cosmic").
	Username string `json:"username,omitempty"`
}

// BillingRequest is the request body for querying compute billing.
type BillingRequest struct {
	// InstanceID filters by instance ID.
	InstanceID string `json:"instance_id,omitempty"`

	// StartDate is the billing period start (ISO 8601).
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the billing period end (ISO 8601).
	EndDate string `json:"end_date,omitempty"`
}

// BillingEntry is a single billing line item.
type BillingEntry struct {
	// InstanceID is the instance identifier.
	InstanceID string `json:"instance_id"`

	// InstanceName is the instance name.
	InstanceName string `json:"instance_name,omitempty"`

	// CostUSD is the total cost in US dollars.
	CostUSD float64 `json:"cost_usd"`

	// UsageHours is the usage duration in hours.
	UsageHours float64 `json:"usage_hours,omitempty"`

	// SKUDescription is the SKU description (e.g. "N1 Predefined Instance Core").
	SKUDescription string `json:"sku_description,omitempty"`

	// StartTime is the billing period start.
	StartTime string `json:"start_time,omitempty"`

	// EndTime is the billing period end.
	EndTime string `json:"end_time,omitempty"`
}

// BillingResponse is the response from a compute billing query.
type BillingResponse struct {
	// Entries is the list of billing line items.
	Entries []BillingEntry `json:"entries"`

	// TotalCostUSD is the total cost across all entries.
	TotalCostUSD float64 `json:"total_cost_usd"`
}

// ComputeTemplates returns available machine configurations with pricing.
func (c *Client) ComputeTemplates(ctx context.Context) (*TemplatesResponse, error) {
	var resp TemplatesResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/compute/templates", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ComputeProvision creates a new compute instance from a template.
// A 1-hour minimum is charged immediately.
func (c *Client) ComputeProvision(ctx context.Context, req *ProvisionRequest) (*ProvisionResponse, error) {
	var resp ProvisionResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/compute/provision", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ComputeInstances lists all compute instances for the authenticated user.
func (c *Client) ComputeInstances(ctx context.Context) (*InstancesResponse, error) {
	var resp InstancesResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/compute/instances", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ComputeInstance returns the full status of a single compute instance.
func (c *Client) ComputeInstance(ctx context.Context, id string) (*InstanceResponse, error) {
	var resp InstanceResponse
	_, err := c.doJSON(ctx, "GET", fmt.Sprintf("/qai/v1/compute/instance/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ComputeDelete tears down a compute instance and finalizes billing.
func (c *Client) ComputeDelete(ctx context.Context, id string) (*DeleteResponse, error) {
	var resp DeleteResponse
	_, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/qai/v1/compute/instance/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ComputeSSHKey injects an SSH public key into a running compute instance.
func (c *Client) ComputeSSHKey(ctx context.Context, id string, req *SSHKeyRequest) error {
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/compute/instance/%s/ssh-key", id), req, nil)
	return err
}

// ComputeKeepalive resets the inactivity timer on a running compute instance.
func (c *Client) ComputeKeepalive(ctx context.Context, id string) error {
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/compute/instance/%s/keepalive", id), nil, nil)
	return err
}

// ComputeBilling queries compute billing from BigQuery via the QAI backend.
func (c *Client) ComputeBilling(ctx context.Context, req *BillingRequest) (*BillingResponse, error) {
	var resp BillingResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/compute/billing", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
