package provider

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func NewTestCheckOutput(name string, value interface{}) func(*terraform.State) error {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			if value == nil {
				return nil
			}
			return fmt.Errorf("Not found: %s", name)
		}

		if !reflect.DeepEqual(rs.Value, value) {
			return fmt.Errorf(
				"output '%s': expected %#v, got %#v",
				name,
				value,
				rs.Value)
		}

		return nil
	}
}

func TestAccEvalFunction_basic(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "output" {
					value = provider::starlark::eval(
						<<-EOT
						result = a + b
						EOT
						,
						{
							a = 10
							b = 20
						}
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("output", "30"),
				),
			},
		},
	})
}

func TestAccEvalFunction_complex(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "result_list" {
					value = provider::starlark::eval(
						<<-EOT
						def get_list():
							return [1, 2, 3]
						
						result = get_list()
						EOT
						,
						{}
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					NewTestCheckOutput("result_list", []interface{}{json.Number("1"), json.Number("2"), json.Number("3")}),
				),
			},
		},
	})
}

func TestAccEvalFunction_return_types(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "ret_string" { 
					value = provider::starlark::eval("result = 'string_val'", {}) 
				}
				output "ret_bool" { 
					value = provider::starlark::eval("result = True", {}) 
				}
				output "ret_int" { 
					value = provider::starlark::eval("result = 100", {}) 
				}
				output "ret_float" { 
					value = provider::starlark::eval("result = 1.5", {}) 
				}
				output "ret_tuple" { 
					value = provider::starlark::eval("result = [1, 'two']", {}) 
				}
				output "ret_object" { 
					value = provider::starlark::eval("result = {'a': 'b'}", {}) 
				}
				output "ret_null" { 
					value = provider::starlark::eval("result = None", {}) 
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					NewTestCheckOutput("ret_string", "string_val"),
					NewTestCheckOutput("ret_bool", "true"), // TF might return bool or string, resource.TestCheckOutput uses string, NewTestCheckOutput checks raw value
					NewTestCheckOutput("ret_int", "100"),
					NewTestCheckOutput("ret_float", "1.5"),
					NewTestCheckOutput("ret_tuple", []interface{}{json.Number("1"), "two"}),
					NewTestCheckOutput("ret_object", map[string]interface{}{
						"a": "b",
					}),
					NewTestCheckOutput("ret_null", nil),
				),
			},
		},
	})
}

func TestAccEvalFunction_input_types(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "type_int_hcl" { 
					value = provider::starlark::eval("result = type(v)", { v = 1 }) 
				}
				output "val_string" {
					value = provider::starlark::eval("result = v", { v = "my_string" })
				}
				output "val_bool" {
					value = provider::starlark::eval("result = v", { v = true })
				}
				output "val_list" {
					value = provider::starlark::eval("result = v", { v = ["a", "b"] })
				}
				output "val_map_key" {
					value = provider::starlark::eval("result = v['key']", { v = { key = "value" } })
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					NewTestCheckOutput("type_int_hcl", "float"), // TF numbers are float/big.Float in dynamic
					NewTestCheckOutput("val_string", "my_string"),
					NewTestCheckOutput("val_bool", "true"),
					NewTestCheckOutput("val_list", []interface{}{"a", "b"}),
					NewTestCheckOutput("val_map_key", "value"),
				),
			},
		},
	})
}

func TestAccEvalFunction_date_time_to_epoch(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "epoch_val" {
					value = provider::starlark::eval(
						<<-EOT
						def is_leap(y):
							return (y % 4 == 0 and y % 100 != 0) or (y % 400 == 0)

						def date_to_epoch(iso):
							# Format: 2023-01-01T00:00:00Z
							y = int(iso[0:4])
							m = int(iso[5:7])
							d = int(iso[8:10])
							h = int(iso[11:13])
							min = int(iso[14:16])
							s = int(iso[17:19])
							
							days = 0
							for cy in range(1970, y):
								days += 366 if is_leap(cy) else 365
							
							md = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
							if is_leap(y):
								md[1] = 29
								
							for cm in range(0, m-1):
								days += md[cm]
								
							days += d - 1
							return days * 86400 + h * 3600 + min * 60 + s

						result = str(date_to_epoch(v))
						EOT
						,
						{ v = "2023-01-01T00:00:00Z" }
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("epoch_val", "1672531200"),
				),
			},
		},
	})
}

func TestAccEvalFunction_date_time_from_epoch(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "date_val" {
					value = provider::starlark::eval(
						<<-EOT
						def is_leap(y):
							return (y % 4 == 0 and y % 100 != 0) or (y % 400 == 0)

						def epoch_to_date(ts):
							ts = int(ts)
							days = ts // 86400
							rem = ts % 86400
							h = rem // 3600
							rem = rem % 3600
							min = rem // 60
							s = rem % 60
							
							y = 1970
							for i in range(1970, 3000):
								d_in_y = 366 if is_leap(i) else 365
								if days < d_in_y:
									y = i
									break
								days -= d_in_y
								
							md = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
							if is_leap(y):
								md[1] = 29
								
							m = 1
							for d_in_m in md:
								if days < d_in_m:
									break
								days -= d_in_m
								m += 1
								
							d = days + 1
							
							# formatting helper
							def pad(n):
								return "0" + str(n) if n < 10 else str(n)
								
							return str(y) + "-" + pad(m) + "-" + pad(d) + "T" + pad(h) + ":" + pad(min) + ":" + pad(s) + "Z"

						result = epoch_to_date(v)
						EOT
						,
						{ v = 1672531200 }
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("date_val", "2023-01-01T00:00:00Z"),
				),
			},
		},
	})
}

func TestAccEvalFunction_parse_cidr(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "cidr_obj" {
					value = provider::starlark::eval(
						<<-EOT
						def ip_to_int(ip):
							parts = ip.split('.')
							return (int(parts[0]) << 24) + (int(parts[1]) << 16) + (int(parts[2]) << 8) + int(parts[3])

						def int_to_ip(val):
							return str((val >> 24) & 0xFF) + "." + str((val >> 16) & 0xFF) + "." + str((val >> 8) & 0xFF) + "." + str(val & 0xFF)

						def parse_cidr(cidr_str):
							parts = cidr_str.split('/')
							ip_int = ip_to_int(parts[0])
							prefix = int(parts[1])
							
							wildcard = (1 << (32 - prefix)) - 1
							mask = ~wildcard & 0xFFFFFFFF
							
							network_int = ip_int & mask
							broadcast_int = network_int | wildcard
							first_int = network_int + 1
							last_int = broadcast_int - 1
							
							return {
								"network": int_to_ip(network_int),
								"netmask": int_to_ip(mask),
								"broadcast": int_to_ip(broadcast_int),
								"firstUsable": int_to_ip(first_int),
								"lastUsable": int_to_ip(last_int),
							}

						result = parse_cidr(v)
						EOT
						,
						{ v = "10.0.0.0/24" }
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					NewTestCheckOutput("cidr_obj", map[string]interface{}{
						"network":     "10.0.0.0",
						"netmask":     "255.255.255.0",
						"broadcast":   "10.0.0.255",
						"firstUsable": "10.0.0.1",
						"lastUsable":  "10.0.0.254",
					}),
				),
			},
		},
	})
}

func TestAccEvalFunction_recursion(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "factorial" {
					value = provider::starlark::eval(
						<<-EOT
						def factorial(n):
							if n == 0:
								return 1
							return n * factorial(n - 1)
						
						result = factorial(5)
						EOT
						,
						{}
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("factorial", "120"),
				),
			},
		},
	})
}

func TestAccEvalFunction_while_loop(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "loop_result" {
					value = provider::starlark::eval(
						<<-EOT
						def run_loop():
							x = 0
							while x < 10:
								x = x + 1
							return x
						
						result = run_loop()
						EOT
						,
						{}
					)
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("loop_result", "10"),
				),
			},
		},
	})
}
