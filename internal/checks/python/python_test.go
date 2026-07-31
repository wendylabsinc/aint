// internal/checks/python/python_test.go
package python_test

import (
	"testing"

	"aint/internal/checks/python"
)

func TestShellTrueDetectsShellInjectionRisk(t *testing.T) {
	findings := python.ShellTrue.Run("script.py", []byte(`subprocess.run(cmd, shell=True)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestShellTrueIgnoresSafeCall(t *testing.T) {
	findings := python.ShellTrue.Run("script.py", []byte(`subprocess.run(cmd)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestEvalExecDetectsEval(t *testing.T) {
	findings := python.EvalExec.Run("script.py", []byte(`eval(user_input)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestEvalExecDetectsExec(t *testing.T) {
	findings := python.EvalExec.Run("script.py", []byte(`exec(user_input)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestEvalExecIgnoresSimilarIdentifier(t *testing.T) {
	findings := python.EvalExec.Run("script.py", []byte(`evaluate(user_input)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestOSSystemDetectsCall(t *testing.T) {
	findings := python.OSSystem.Run("script.py", []byte(`os.system(cmd)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestOSSystemIgnoresUnrelatedOSCall(t *testing.T) {
	findings := python.OSSystem.Run("script.py", []byte(`os.path.join(a, b)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPickleLoadDetectsLoads(t *testing.T) {
	findings := python.PickleLoad.Run("script.py", []byte(`data = pickle.loads(payload)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPickleLoadIgnoresJSONLoads(t *testing.T) {
	findings := python.PickleLoad.Run("script.py", []byte(`data = json.loads(payload)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestYAMLUnsafeLoadDetectsLoad(t *testing.T) {
	findings := python.YAMLUnsafeLoad.Run("script.py", []byte(`config = yaml.load(f)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestYAMLUnsafeLoadIgnoresSafeLoad(t *testing.T) {
	findings := python.YAMLUnsafeLoad.Run("script.py", []byte(`config = yaml.safe_load(f)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestRequestsNoVerifyDetectsVerifyFalse(t *testing.T) {
	findings := python.RequestsNoVerify.Run("script.py", []byte(`requests.get(url, verify=False)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestRequestsNoVerifyIgnoresDefaultVerify(t *testing.T) {
	findings := python.RequestsNoVerify.Run("script.py", []byte(`requests.get(url)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestWeakHashDetectsMD5(t *testing.T) {
	findings := python.WeakHash.Run("script.py", []byte(`hashlib.md5(password.encode())`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestWeakHashIgnoresSHA256(t *testing.T) {
	findings := python.WeakHash.Run("script.py", []byte(`hashlib.sha256(password.encode())`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestAssertSecurityCheckDetectsIsAdmin(t *testing.T) {
	findings := python.AssertSecurityCheck.Run("script.py", []byte(`assert user.is_admin`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestAssertSecurityCheckIgnoresGenericAssert(t *testing.T) {
	findings := python.AssertSecurityCheck.Run("script.py", []byte(`assert len(items) > 0`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
