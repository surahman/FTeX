package config_loader

// configTestData will return a map of test data containing valid and invalid test configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,
		"valid": `
parent_level_1:
  child_one: child_one_key
  child_two: 999
  child_three: true
parent_level_2:
  child_one: 5
  child_two: true
  child_three: "abcdef"`,
		"valid - required": `
parent_level_1:
  child_one: child_one_key
  child_two: 999
parent_level_2:
  child_one: 5
  child_three: "abcdef"`,
		"invalid - no lvl 1": `
parent_level_2:
  child_one: 5
  child_two: true
  child_three: "abcdef"`,
		"invalid - no lvl1 child1": `
parent_level_1:
  child_two: 999
  child_three: true
parent_level_2:
  child_one: 5
  child_two: true
  child_three: "abcdef"`,
		"invalid - lvl1 child1 below threshold": `
parent_level_1:
  child_one: child_one_key
  child_two: 9
  child_three: true
parent_level_2:
  child_one: 5
  child_two: true
  child_three: "abcdef"`,
		"invalid - lvl2 no child1": `
parent_level_1:
  child_one: child_one_key
  child_two: 999
  child_three: true
parent_level_2:
  child_two: true
  child_three: "abcdef"`,
		"invalid - lvl2 child1 below threshold": `
parent_level_1:
  child_one: child_one_key
  child_two: 999
  child_three: true
parent_level_2:
  child_one: 2
  child_two: true
  child_three: "abcdef"`,
		"invalid - lvl2 no child3": `
parent_level_1:
  child_one: child_one_key
  child_two: 999
  child_three: true
parent_level_2:
  child_one: 5
  child_two: true`,
		"invalid - lvl2 child3 below threshold": `
parent_level_1:
  child_one: child_one_key
  child_two: 999
  child_three: true
parent_level_2:
  child_one: 5
  child_two: true
  child_three: "ab"`,
	}
}
