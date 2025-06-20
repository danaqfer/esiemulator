package propertymanager

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewPropertyManager(t *testing.T) {
	pm := NewPropertyManager(true)
	if pm == nil {
		t.Fatal("NewPropertyManager returned nil")
	}
	if !pm.Debug {
		t.Error("Debug should be true")
	}
	if pm.Rules == nil {
		t.Error("Rules map should be initialized")
	}
	if pm.Behaviors == nil {
		t.Error("Behaviors map should be initialized")
	}
	if pm.Variables == nil {
		t.Error("Variables map should be initialized")
	}
}

func TestLoadProperty(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/test"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Test"/>
					<option name="value" value="test-value"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
	<behaviors>
		<behavior name="test-behavior">
			<option name="test_option" value="test_value"/>
		</behavior>
	</behaviors>
	<variables>
		<variable name="test_var" value="test_value"/>
	</variables>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	if pm.Property == nil {
		t.Fatal("Property should be loaded")
	}
	if pm.Property.Name != "test-property" {
		t.Errorf("Expected property name 'test-property', got '%s'", pm.Property.Name)
	}
	if len(pm.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(pm.Rules))
	}
	if len(pm.Behaviors) != 1 {
		t.Errorf("Expected 1 behavior, got %d", len(pm.Behaviors))
	}
	if len(pm.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(pm.Variables))
	}
}

func TestProcessRequest_BasicRule(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/test"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Test"/>
					<option name="value" value="test-value"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.MatchedRules[0] != "test-rule" {
		t.Errorf("Expected matched rule 'test-rule', got '%s'", result.MatchedRules[0])
	}
	if len(result.ExecutedBehaviors) != 1 {
		t.Errorf("Expected 1 executed behavior, got %d", len(result.ExecutedBehaviors))
	}
	if result.ModifiedHeaders["X-Test"] != "test-value" {
		t.Errorf("Expected header X-Test=test-value, got '%s'", result.ModifiedHeaders["X-Test"])
	}
}

func TestProcessRequest_NoMatch(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/test"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Test"/>
					<option name="value" value="test-value"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/other", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 0 {
		t.Errorf("Expected 0 matched rules, got %d", len(result.MatchedRules))
	}
	if len(result.ExecutedBehaviors) != 0 {
		t.Errorf("Expected 0 executed behaviors, got %d", len(result.ExecutedBehaviors))
	}
}

func TestProcessRequest_MultipleCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/test"/>
			<criteria name="method" option="equals" value="GET"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Test"/>
					<option name="value" value="test-value"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
}

func TestProcessRequest_HeaderCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="header" option="User-Agent" extract="contains" value="Mozilla"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Browser"/>
					<option name="value" value="browser"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Browser"] != "browser" {
		t.Errorf("Expected header X-Browser=browser, got '%s'", result.ModifiedHeaders["X-Browser"])
	}
}

func TestProcessRequest_CookieCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="cookie" option="session" extract="equals" value="abc123"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Session"/>
					<option name="value" value="valid"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Session"] != "valid" {
		t.Errorf("Expected header X-Session=valid, got '%s'", result.ModifiedHeaders["X-Session"])
	}
}

func TestProcessRequest_RedirectBehavior(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/old"/>
			<behaviors>
				<behavior name="redirect">
					<option name="destination" value="/new"/>
					<option name="status_code" value="301"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/old", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["Location"] != "/new" {
		t.Errorf("Expected Location header /new, got '%s'", result.ModifiedHeaders["Location"])
	}
	if result.ModifiedHeaders["Status"] != "301" {
		t.Errorf("Expected Status header 301, got '%s'", result.ModifiedHeaders["Status"])
	}
	if !strings.Contains(result.ResponseContent, "/new") {
		t.Error("Response content should contain redirect URL")
	}
}

func TestProcessRequest_SetVariableBehavior(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/test"/>
			<behaviors>
				<behavior name="set_variable">
					<option name="variable_name" value="test_var"/>
					<option name="value" value="test_value"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.Variables["test_var"] != "test_value" {
		t.Errorf("Expected variable test_var=test_value, got '%s'", result.Variables["test_var"])
	}
}

func TestProcessRequest_GzipBehavior(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="ends_with" value=".html"/>
			<behaviors>
				<behavior name="gzip_response">
					<option name="enabled" value="true"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test.html", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["Content-Encoding"] != "gzip" {
		t.Errorf("Expected Content-Encoding header gzip, got '%s'", result.ModifiedHeaders["Content-Encoding"])
	}
}

func TestProcessRequest_ChildRules(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="parent-rule">
			<criteria name="path" option="starts_with" value="/api"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-API"/>
					<option name="value" value="true"/>
				</behavior>
			</behaviors>
			<children>
				<rule name="child-rule">
					<criteria name="path" option="equals" value="/api/v1"/>
					<behaviors>
						<behavior name="set_response_header">
							<option name="header_name" value="X-Version"/>
							<option name="value" value="v1"/>
						</behavior>
					</behaviors>
				</rule>
			</children>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/v1", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 2 {
		t.Errorf("Expected 2 matched rules, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-API"] != "true" {
		t.Errorf("Expected header X-API=true, got '%s'", result.ModifiedHeaders["X-API"])
	}
	if result.ModifiedHeaders["X-Version"] != "v1" {
		t.Errorf("Expected header X-Version=v1, got '%s'", result.ModifiedHeaders["X-Version"])
	}
}

func TestProcessRequest_QueryCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="query" option="contains" value="debug=true"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Debug"/>
					<option name="value" value="enabled"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test?debug=true&other=value", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Debug"] != "enabled" {
		t.Errorf("Expected header X-Debug=enabled, got '%s'", result.ModifiedHeaders["X-Debug"])
	}
}

func TestProcessRequest_HostCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="host" option="equals" value="example.com"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Host"/>
					<option name="value" value="example"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Host = "example.com"
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Host"] != "example" {
		t.Errorf("Expected header X-Host=example, got '%s'", result.ModifiedHeaders["X-Host"])
	}
}

func TestProcessRequest_UserAgentCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="user_agent" option="contains" value="Mobile"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Device"/>
					<option name="value" value="mobile"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Device"] != "mobile" {
		t.Errorf("Expected header X-Device=mobile, got '%s'", result.ModifiedHeaders["X-Device"])
	}
}

func TestProcessRequest_ClientIPCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="client_ip" option="starts_with" value="192.168"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Network"/>
					<option name="value" value="internal"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Network"] != "internal" {
		t.Errorf("Expected header X-Network=internal, got '%s'", result.ModifiedHeaders["X-Network"])
	}
}

func TestProcessRequest_RegexPathCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="regex" value="/api/v[0-9]+/.*"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-API-Version"/>
					<option name="value" value="detected"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/v2/users", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-API-Version"] != "detected" {
		t.Errorf("Expected header X-API-Version=detected, got '%s'", result.ModifiedHeaders["X-API-Version"])
	}
}

func TestProcessRequest_MultipleBehaviors(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<criteria name="path" option="equals" value="/test"/>
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Header1"/>
					<option name="value" value="value1"/>
				</behavior>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Header2"/>
					<option name="value" value="value2"/>
				</behavior>
				<behavior name="set_variable">
					<option name="variable_name" value="test_var"/>
					<option name="value" value="test_value"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if len(result.ExecutedBehaviors) != 3 {
		t.Errorf("Expected 3 executed behaviors, got %d", len(result.ExecutedBehaviors))
	}
	if result.ModifiedHeaders["X-Header1"] != "value1" {
		t.Errorf("Expected header X-Header1=value1, got '%s'", result.ModifiedHeaders["X-Header1"])
	}
	if result.ModifiedHeaders["X-Header2"] != "value2" {
		t.Errorf("Expected header X-Header2=value2, got '%s'", result.ModifiedHeaders["X-Header2"])
	}
	if result.Variables["test_var"] != "test_value" {
		t.Errorf("Expected variable test_var=test_value, got '%s'", result.Variables["test_var"])
	}
}

func TestProcessRequest_NoCriteria(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
	<rules>
		<rule name="test-rule">
			<behaviors>
				<behavior name="set_response_header">
					<option name="header_name" value="X-Default"/>
					<option name="value" value="default"/>
				</behavior>
			</behaviors>
		</rule>
	</rules>
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/any-path", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
	if result.ModifiedHeaders["X-Default"] != "default" {
		t.Errorf("Expected header X-Default=default, got '%s'", result.ModifiedHeaders["X-Default"])
	}
}

func TestProcessRequest_InvalidXML(t *testing.T) {
	xmlData := []byte(`invalid xml content`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err == nil {
		t.Fatal("LoadProperty should fail with invalid XML")
	}
}

func TestProcessRequest_EmptyProperty(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<property name="test-property" version="1">
</property>`)

	pm := NewPropertyManager(false)
	err := pm.LoadProperty(xmlData)
	if err != nil {
		t.Fatalf("LoadProperty failed: %v", err)
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	result, err := pm.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if len(result.MatchedRules) != 0 {
		t.Errorf("Expected 0 matched rules, got %d", len(result.MatchedRules))
	}
	if len(result.ExecutedBehaviors) != 0 {
		t.Errorf("Expected 0 executed behaviors, got %d", len(result.ExecutedBehaviors))
	}
}
