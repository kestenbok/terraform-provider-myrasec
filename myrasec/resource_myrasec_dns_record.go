package myrasec

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Myra-Security-GmbH/myrasec-go"
	"github.com/Myra-Security-GmbH/myrasec-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

//
// resourceMyrasecDNSRecord ...
//
func resourceMyrasecDNSRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceMyrasecDNSRecordCreate,
		Read:   resourceMyrasecDNSRecordRead,
		Delete: resourceMyrasecDNSRecordDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(i interface{}) string {
					return strings.ToLower(i.(string))
				},
				Description: "The Domain for the DNS record.",
			},
			"record_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the DNS record.",
			},
			"modified": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of last modification.",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of creation.",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(i interface{}) string {
					return strings.ToLower(i.(string))
				},
				Description: "Subdomain name of a DNS record.",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Time to live.",
			},
			"record_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"A", "AAAA", "MX", "CNAME", "TXT", "NS", "SRV", "CAA"}, false),
				Description:  "A record type to identify the type of a record. Valid types are: A, AAAA, MX, CNAME, TXT, NS, SRV and CAA.",
			},
			"alternative_cname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The alternative CNAME that points to the record.",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Define wether this subdomain should be protected by Myra or not.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Define wether this DNS record is enabled or not.",
			},
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
				Description: "A comment to describe this DNS record.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Depends on the record type. Typically an IPv4/6 address or a domain entry.",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Priority of MX records.",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Port for SRV records.",
			},
			"upstream_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"upstream_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the upstream configuration.",
						},
						"modified": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date of last modification.",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date of creation.",
						},
						"backup": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							ForceNew:    true,
							Description: "Marks the server as a backup server. It will be used when the primary servers are unavailable. Cannot be used in combination with \"Preserve client IP on the same upstream\".",
						},
						"down": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							ForceNew:    true,
							Description: "Marks the server as unavailable.",
						},
						"fail_timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
							ForceNew:    true,
							Description: "Double usage: 1. Time period in which the max_fails must occur until the upstream is deactivated. 2. Time period the upstream is deactivated until it is reactivated. The time during which the specified number of unsuccessful attempts \"Max fails\" to communicate with the server should happen to consider the server unavailable. Also the period of time the server will be considered unavailable. Default is 10 seconds.",
						},
						"max_fails": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							ForceNew:    true,
							Description: "The number of unsuccessful attempts to communicate with the server that should happen in the duration set by \"Fail timeout\" to consider the server unavailable. Also the server is considered unavailable for the duration set by \"Fail timeout\". By default, the number of unsuccessful attempts is set to 1. Setting the value to zero disables the accounting of attempts. What is considered an unsuccessful attempt is defined by the \"Next upstream error handling\".",
						},
						"weight": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
							ForceNew:    true,
							Description: "Weight defines the count of requests a upstream handles before the next upstream is selected.",
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

//
// resourceMyrasecDNSRecordCreate ...
//
func resourceMyrasecDNSRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*myrasec.API)

	record, err := buildDNSRecord(d, meta)
	if err != nil {
		return fmt.Errorf("Error building DNS record: %s", err)
	}

	resp, err := client.CreateDNSRecord(record, d.Get("domain_name").(string))
	if err != nil {
		return fmt.Errorf("Error creating DNS record: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", resp.ID))
	return resourceMyrasecDNSRecordRead(d, meta)
}

//
// resourceMyrasecDNSRecordRead ...
//
func resourceMyrasecDNSRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*myrasec.API)

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing record id: %s", err)
	}

	records, err := client.ListDNSRecords(d.Get("domain_name").(string), map[string]string{"loadbalancer": "true"})
	if err != nil {
		return fmt.Errorf("Error fetching DNS records: %s", err)
	}

	for _, r := range records {
		if r.ID != recordID {
			continue
		}
		d.Set("record_id", r.ID)
		d.Set("name", r.Name)
		d.Set("value", r.Value)
		d.Set("record_type", r.RecordType)
		d.Set("ttl", r.TTL)
		d.Set("alternative_cname", r.AlternativeCNAME)
		d.Set("active", r.Active)
		d.Set("enabled", r.Enabled)
		d.Set("priority", r.Priority)
		d.Set("port", r.Port)
		d.Set("created", r.Created.Format(time.RFC3339))
		d.Set("modified", r.Modified.Format(time.RFC3339))
		d.Set("comment", r.Comment)

		if r.UpstreamOptions != nil {
			d.Set("upstream_options", map[string]interface{}{
				"upstream_id":  r.UpstreamOptions.ID,
				"created":      r.UpstreamOptions.Created.Format(time.RFC3339),
				"modified":     r.UpstreamOptions.Modified.Format(time.RFC3339),
				"backup":       r.UpstreamOptions.Backup,
				"down":         r.UpstreamOptions.Down,
				"fail_timeout": r.UpstreamOptions.FailTimeout,
				"max_fails":    r.UpstreamOptions.MaxFails,
				"weight":       r.UpstreamOptions.Weight,
			})
		}
		break
	}

	return nil
}

//
// resourceMyrasecDNSRecordDelete ...
//
func resourceMyrasecDNSRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*myrasec.API)

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing record id: %s", err)
	}

	log.Printf("[INFO] Deleting DNS record: %v", recordID)

	record, err := buildDNSRecord(d, meta)
	if err != nil {
		return fmt.Errorf("Error building DNS record: %s", err)
	}

	_, err = client.DeleteDNSRecord(record, d.Get("domain_name").(string))
	if err != nil {
		return fmt.Errorf("Error deleting DNS record: %s", err)
	}
	return nil
}

//
// buildDNSRecord ...
//
func buildDNSRecord(d *schema.ResourceData, meta interface{}) (*myrasec.DNSRecord, error) {
	record := &myrasec.DNSRecord{
		Name:             d.Get("name").(string),
		Value:            d.Get("value").(string),
		RecordType:       d.Get("record_type").(string),
		TTL:              d.Get("ttl").(int),
		AlternativeCNAME: d.Get("alternative_cname").(string),
		Active:           d.Get("active").(bool),
		Enabled:          d.Get("enabled").(bool),
		Priority:         d.Get("priority").(int),
		Port:             d.Get("port").(int),
		Comment:          d.Get("comment").(string),
	}

	if d.Get("record_id").(int) > 0 {
		record.ID = d.Get("record_id").(int)
	}

	if len(d.Get("created").(string)) > 0 {
		created, err := time.Parse(time.RFC3339, d.Get("created").(string))
		if err != nil {
			return nil, err
		}

		record.Created = &types.DateTime{
			Time: created,
		}
	}

	if len(d.Get("modified").(string)) > 0 {
		modified, err := time.Parse(time.RFC3339, d.Get("modified").(string))
		if err != nil {
			return nil, err
		}
		record.Modified = &types.DateTime{
			Time: modified,
		}
	}

	options, ok := d.GetOk("upstream_options")
	if !ok {
		return record, nil
	}

	for _, upstream := range options.([]interface{}) {
		opts, err := buildUpstreamOptions(upstream)
		if err != nil {
			return nil, err
		}

		record.UpstreamOptions = opts
	}

	return record, nil
}

//
// buildUpstreamOptions ...
//
func buildUpstreamOptions(upstream interface{}) (*myrasec.UpstreamOptions, error) {
	options := &myrasec.UpstreamOptions{}

	for key, val := range upstream.(map[string]interface{}) {
		switch key {
		case "upstream_id":
			options.ID = val.(int)
		case "modified":
			if len(val.(string)) <= 0 {
				continue
			}
			modified, err := time.Parse(time.RFC3339, val.(string))
			if err != nil {
				return nil, err
			}
			options.Modified = &types.DateTime{
				Time: modified,
			}
		case "created":
			if len(val.(string)) <= 0 {
				continue
			}
			created, err := time.Parse(time.RFC3339, val.(string))
			if err != nil {
				return nil, err
			}
			options.Created = &types.DateTime{
				Time: created,
			}
		case "backup":
			options.Backup = val.(bool)
		case "down":
			options.Down = val.(bool)
		case "fail_timeout":
			options.FailTimeout = val.(int)
		case "max_fails":
			options.MaxFails = val.(int)
		case "weight":
			options.Weight = val.(int)
		}
	}

	return options, nil
}
