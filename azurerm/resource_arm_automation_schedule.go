package azurerm

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/automation/mgmt/2015-10-31/automation"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/set"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmAutomationSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmAutomationScheduleCreateUpdate,
		Read:   resourceArmAutomationScheduleRead,
		Update: resourceArmAutomationScheduleCreateUpdate,
		Delete: resourceArmAutomationScheduleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[^<>*%&:\\?.+/]{0,127}[^<>*%&:\\?.+/\s]$`),
					`The name length must be from 1 to 128 characters. The name cannot contain special characters < > * % & : \ ? . + / and cannot end with a whitespace character.`,
				),
			},

			"resource_group_name": resourceGroupNameSchema(),

			//this is AutomationAccountName in the SDK
			"account_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Deprecated:    "account_name has been renamed to automation_account_name for clarity and to match the azure API",
				ConflictsWith: []string{"automation_account_name"},
			},

			"automation_account_name": {
				Type:     schema.TypeString,
				Optional: true, //todo change to required once account_name has been removed
				Computed: true,
				//ForceNew:      true, //todo this needs to come back once account_name has been removed
				ConflictsWith: []string{"account_name"},
			},

			"frequency": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(automation.Day),
					string(automation.Hour),
					string(automation.Month),
					string(automation.OneTime),
					string(automation.Week),
				}, true),
			},

			//ignored when frequency is `OneTime`
			"interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true, //defaults to 1 if frequency is not OneTime
				ValidateFunc: validation.IntBetween(1, 100),
			},

			"start_time": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppress.RFC3339Time,
				ValidateFunc:     validate.RFC3339DateInFutureBy(time.Duration(5) * time.Minute),
				//defaults to now + 7 minutes in create function if not set
			},

			"expiry_time": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true, //same as start time when OneTime, ridiculous value when recurring: "9999-12-31T15:59:00-08:00"
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validate.RFC3339Time,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UTC",
				//todo figure out how to validate this properly
			},

			"advanced_schedule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"week_days": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									string(automation.Monday),
									string(automation.Tuesday),
									string(automation.Wednesday),
									string(automation.Thursday),
									string(automation.Friday),
									string(automation.Saturday),
									string(automation.Sunday),
								}, true),
							},
							Set: set.HashStringIgnoreCase,
						},
						"month_days": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeInt,
								ValidateFunc: validate.IntBetweenAndNot(-1, 31, 0),
							},
							Set: set.HashInt,
						},

						"monthly_occurrence": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day": {
										Type:             schema.TypeString,
										Required:         true,
										DiffSuppressFunc: suppress.CaseDifference,
										ValidateFunc: validation.StringInSlice([]string{
											string(automation.Monday),
											string(automation.Tuesday),
											string(automation.Wednesday),
											string(automation.Thursday),
											string(automation.Friday),
											string(automation.Saturday),
											string(automation.Sunday),
										}, true),
									},
									"occurrence": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 5),
									},
								},
							},
						},
					},
				},
			},
		},

		CustomizeDiff: func(diff *schema.ResourceDiff, v interface{}) error {

			frequency := strings.ToLower(diff.Get("frequency").(string))
			interval, _ := diff.GetOk("interval")
			if frequency == "onetime" && interval.(int) > 0 {
				return fmt.Errorf("`interval` cannot be set when frequency is not OneTime")
			}

			advancedSchedules, hasAdvancedSchedule := diff.GetOk("advanced_schedule")
			if hasAdvancedSchedule {
				if asl := advancedSchedules.([]interface{}); len(asl) > 0 {
					if frequency != "week" && frequency != "month" {
						return fmt.Errorf("`advanced_schedule` can only be set when frequency is `Week` or `Month`")
					}

					as := asl[0].(map[string]interface{})
					if frequency == "week" && as["week_days"].(*schema.Set).Len() == 0 {
						return fmt.Errorf("`week_days` must be set when frequency is `Week`")
					}
					if frequency == "month" && as["month_days"].(*schema.Set).Len() == 0 && len(as["monthly_occurrence"].([]interface{})) == 0 {
						return fmt.Errorf("Either `month_days` or `monthly_occurrence` must be set when frequency is `Month`")
					}
				}
			}

			_, hasAccount := diff.GetOk("automation_account_name")
			_, hasAutomationAccountWeb := diff.GetOk("account_name")
			if !hasAccount && !hasAutomationAccountWeb {
				return fmt.Errorf("`automation_account_name` must be set")
			}

			//if automation_account_name changed or account_name changed to or from nil force a new resource
			//remove once we remove the deprecated property
			oAan, nAan := diff.GetChange("automation_account_name")
			if oAan != "" && nAan != "" {
				diff.ForceNew("automation_account_name")
			}

			oAn, nAn := diff.GetChange("account_name")
			if oAn != "" && nAn != "" {
				diff.ForceNew("account_name")
			}

			return nil
		},
	}
}

func resourceArmAutomationScheduleCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationScheduleClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM Automation Schedule creation.")

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	frequency := d.Get("frequency").(string)

	timeZone := d.Get("timezone").(string)
	description := d.Get("description").(string)

	//CustomizeDiff should ensure one of these two is set
	//todo remove this once `account_name` is removed
	accountName := ""
	if v, ok := d.GetOk("automation_account_name"); ok {
		accountName = v.(string)
	} else if v, ok := d.GetOk("account_name"); ok {
		accountName = v.(string)
	}

	parameters := automation.ScheduleCreateOrUpdateParameters{
		Name: &name,
		ScheduleCreateOrUpdateProperties: &automation.ScheduleCreateOrUpdateProperties{
			Description: &description,
			Frequency:   automation.ScheduleFrequency(frequency),
			TimeZone:    &timeZone,
		},
	}
	properties := parameters.ScheduleCreateOrUpdateProperties

	//start time can default to now + 7 (5 could be invalid by the time the API is called)
	if v, ok := d.GetOk("start_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string)) //should be validated by the schema
		properties.StartTime = &date.Time{Time: t}
	} else {
		properties.StartTime = &date.Time{Time: time.Now().Add(time.Duration(7) * time.Minute)}
	}

	if v, ok := d.GetOk("expiry_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string)) //should be validated by the schema
		properties.ExpiryTime = &date.Time{Time: t}
	}

	//only pay attention to interval if frequency is not OneTime, and default it to 1 if not set
	if properties.Frequency != automation.OneTime {
		if v, ok := d.GetOk("interval"); ok {
			properties.Interval = utils.Int32(int32(v.(int)))
		} else {
			properties.Interval = 1
		}
	}

	//only pay attention to the advanced schedule if frequency is either Week or Month
	if properties.Frequency == automation.Week || properties.Frequency == automation.Month {
		if v, ok := d.GetOk("advanced_schedule"); ok {
			if vl := v.([]interface{}); len(vl) > 0 {
				advancedRef, err := expandArmAutomationScheduleAdvanced(vl)
				if err != nil {
					return err
				}
				properties.AdvancedSchedule = advancedRef
			}
		}
	}

	_, err := client.CreateOrUpdate(ctx, resGroup, accountName, name, parameters)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resGroup, accountName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Schedule '%s' (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmAutomationScheduleRead(d, meta)
}

func resourceArmAutomationScheduleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationScheduleClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["schedules"]
	resGroup := id.ResourceGroup
	accountName := id.Path["automationAccounts"]

	resp, err := client.Get(ctx, resGroup, accountName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AzureRM Automation Schedule '%s': %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("automation_account_name", accountName)
	d.Set("account_name", accountName) //todo remove once `account_name` is removed
	d.Set("frequency", string(resp.Frequency))

	if v := resp.StartTime; v != nil {
		d.Set("start_time", string(v.Format(time.RFC3339)))
	}
	if v := resp.ExpiryTime; v != nil {
		d.Set("expiry_time", string(v.Format(time.RFC3339)))
	}
	if v := resp.Interval; v != nil {
		//seems to me missing its type in swagger, leading to it being a interface{} float64
		d.Set("interval", int(v.(float64)))
	}
	if v := resp.Description; v != nil {
		d.Set("description", v)
	}
	if v := resp.TimeZone; v != nil {
		d.Set("timezone", v)
	}
	if err := d.Set("advanced_schedule", flattenArmAutomationScheduleAdvanced(resp.AdvancedSchedule)); err != nil {
		return fmt.Errorf("Error setting `advanced_schedule`: %+v", err)
	}
	return nil
}

func resourceArmAutomationScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationScheduleClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["schedules"]
	resGroup := id.ResourceGroup
	accountName := id.Path["automationAccounts"]

	resp, err := client.Delete(ctx, resGroup, accountName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("Error issuing AzureRM delete request for Automation Schedule '%s': %+v", name, err)
		}
	}

	return nil
}

func expandArmAutomationScheduleAdvanced(v []interface{}) (*automation.AdvancedSchedule, error) {

	// Get the one and only advance schedule configuration
	advancedSchedule := v[0].(map[string]interface{})
	weekDays := advancedSchedule["week_days"].(*schema.Set).List()
	monthDays := advancedSchedule["month_days"].(*schema.Set).List()
	monthlyOccurrences := advancedSchedule["monthly_occurrence"].([]interface{})

	expandedAdvancedSchedule := automation.AdvancedSchedule{}

	// If frequency is set to `Month` the `week_days` array cannot be set (even empty), otherwise the API returns an error.
	// Interestingly enough, during update it can be set and it will not return an error.
	if len(weekDays) > 0 {
		expandedWeekDays := make([]string, len(weekDays))
		for i := range weekDays {
			expandedWeekDays[i] = weekDays[i].(string)
		}
		expandedAdvancedSchedule.WeekDays = &expandedWeekDays
	}

	// Same as above with `week_days`
	if len(monthDays) > 0 {
		expandedMonthDays := make([]int32, len(monthDays))
		for i := range monthDays {
			expandedMonthDays[i] = int32(monthDays[i].(int))
		}
		expandedAdvancedSchedule.MonthDays = &expandedMonthDays
	}

	expandedMonthlyOccurrences := make([]automation.AdvancedScheduleMonthlyOccurrence, len(monthlyOccurrences))
	for i := range monthlyOccurrences {
		m := monthlyOccurrences[i].(map[string]interface{})
		occurrence := int32(m["occurrence"].(int))

		expandedMonthlyOccurrences[i] = automation.AdvancedScheduleMonthlyOccurrence{
			Occurrence: &occurrence,
			Day:        automation.ScheduleDay(m["day"].(string)),
		}
	}
	expandedAdvancedSchedule.MonthlyOccurrences = &expandedMonthlyOccurrences

	return &expandedAdvancedSchedule, nil
}

func flattenArmAutomationScheduleAdvanced(v *automation.AdvancedSchedule) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	result := make(map[string]interface{})

	flattenedWeekDays := schema.NewSet(set.HashStringIgnoreCase, []interface{}{})
	if v.WeekDays != nil {
		for i := range *v.WeekDays {
			flattenedWeekDays.Add((*v.WeekDays)[i])
		}
	}
	result["week_days"] = flattenedWeekDays

	flattenedMonthDays := schema.NewSet(set.HashInt, []interface{}{})
	if v.MonthDays != nil {
		for i := range *v.MonthDays {
			flattenedMonthDays.Add(int(((*v.MonthDays)[i])))
		}
	}
	result["month_days"] = flattenedMonthDays

	flattenedMonthlyOccurrences := make([]map[string]interface{}, 0)
	if v.MonthlyOccurrences != nil {
		for i := range *v.MonthlyOccurrences {
			f := make(map[string]interface{})
			f["day"] = (*v.MonthlyOccurrences)[i].Day
			f["occurrence"] = int(*(*v.MonthlyOccurrences)[i].Occurrence)
			flattenedMonthlyOccurrences = append(flattenedMonthlyOccurrences, f)
		}
	}
	result["monthly_occurrence"] = flattenedMonthlyOccurrences

	return []interface{}{result}
}
