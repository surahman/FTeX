package configloader

// configTestData will return a map of test data containing valid and invalid test configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,
		"valid": `
parentLevel1:
  childOne: childOne_key
  childTwo: 999
  childThree: true
parentLevel2:
  childOne: 5
  childTwo: true
  childThree: "abcdef"`,
		"valid - required": `
parentLevel1:
  childOne: childOne_key
  childTwo: 999
parentLevel2:
  childOne: 5
  childThree: "abcdef"`,
		"invalid - no lvl 1": `
parentLevel2:
  childOne: 5
  childTwo: true
  childThree: "abcdef"`,
		"invalid - no lvl1 child1": `
parentLevel1:
  childTwo: 999
  childThree: true
parentLevel2:
  childOne: 5
  childTwo: true
  childThree: "abcdef"`,
		"invalid - lvl1 child1 below threshold": `
parentLevel1:
  childOne: childOne_key
  childTwo: 9
  childThree: true
parentLevel2:
  childOne: 5
  childTwo: true
  childThree: "abcdef"`,
		"invalid - lvl2 no child1": `
parentLevel1:
  childOne: childOne_key
  childTwo: 999
  childThree: true
parentLevel2:
  childTwo: true
  childThree: "abcdef"`,
		"invalid - lvl2 child1 below threshold": `
parentLevel1:
  childOne: childOne_key
  childTwo: 999
  childThree: true
parentLevel2:
  childOne: 2
  childTwo: true
  childThree: "abcdef"`,
		"invalid - lvl2 no child3": `
parentLevel1:
  childOne: childOne_key
  childTwo: 999
  childThree: true
parentLevel2:
  childOne: 5
  childTwo: true`,
		"invalid - lvl2 child3 below threshold": `
parentLevel1:
  childOne: childOne_key
  childTwo: 999
  childThree: true
parentLevel2:
  childOne: 5
  childTwo: true
  childThree: "ab"`,
	}
}
