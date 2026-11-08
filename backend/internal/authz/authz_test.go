package authz

import "testing"

func TestCatalogIsWellFormed(t *testing.T) {
	seen := make(map[string]string)
	for _, group := range Catalog {
		if group.Key == "" || group.Label == "" {
			t.Fatalf("group has empty key/label: %+v", group)
		}
		if len(group.Permissions) == 0 {
			t.Fatalf("group %q has no permissions", group.Key)
		}
		for _, permission := range group.Permissions {
			if permission.Key == "" || permission.Label == "" {
				t.Fatalf("permission has empty key/label: %+v", permission)
			}
			if previous, ok := seen[permission.Key]; ok {
				t.Fatalf("duplicate permission %q in groups %q and %q", permission.Key, previous, group.Key)
			}
			seen[permission.Key] = group.Key
			if !IsValid(permission.Key) {
				t.Fatalf("permission %q is not valid", permission.Key)
			}
		}
	}
	if len(All()) != len(seen) {
		t.Fatalf("All() returned %d keys, catalog has %d", len(All()), len(seen))
	}
}

func TestIsValidRejectsUnknownKey(t *testing.T) {
	if IsValid("problem.nope") {
		t.Fatal("unknown permission key should be rejected")
	}
	if IsValid("") {
		t.Fatal("empty permission key should be rejected")
	}
}
