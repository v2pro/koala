package trace

import (
	"testing"
	"github.com/stretchr/testify/require"
	"fmt"
)

func Test_simplest_function(t *testing.T) {
	should := require.New(t)
	_, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 0, `function hello() {`, func(
			fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
			return fmt.Sprintf("%v();", functionName)
		})
	should.Equal(0, leftIdx)
	should.Equal(18, rightIdx)
	should.Equal("function hello() { hello(); }\nfunction orig_hello() {", substitution)
}

func Test_public_function(t *testing.T) {
	should := require.New(t)
	isDefinedInClass := false
	bodyProvider := func(
		fileName string, lineNo int, isDefinedInClass_ bool, functionName string, argumentsDef string) string {
		isDefinedInClass = isDefinedInClass_
		return fmt.Sprintf("%v();", functionName)
	}
	_, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 0, `public	function	hello()	{`, bodyProvider)
	should.Equal(0, leftIdx)
	should.Equal(25, rightIdx)
	should.True(isDefinedInClass)
	should.Equal("public\tfunction\thello()\t{ hello(); }\npublic\tfunction\torig_hello()\t{", substitution)
	_, leftIdx, rightIdx, substitution = subPhpFunctionDefinition(
		"a.php", 0, `public static function hello() {`, bodyProvider)
	should.Equal(0, leftIdx)
	_, leftIdx, rightIdx, substitution = subPhpFunctionDefinition(
		"a.php", 0, `static public function hello() {`, bodyProvider)
	should.Equal(0, leftIdx)
	_, leftIdx, rightIdx, substitution = subPhpFunctionDefinition(
		"a.php", 0, `private static function hello() {`, bodyProvider)
	should.Equal(0, leftIdx)
	_, leftIdx, rightIdx, substitution = subPhpFunctionDefinition(
		"a.php", 0, `static private function hello() {`, bodyProvider)
	should.Equal(0, leftIdx)
	_, leftIdx, rightIdx, substitution = subPhpFunctionDefinition(
		"a.php", 0, `static function hello() {`, bodyProvider)
	should.Equal(0, leftIdx)
}

func Test_function_with_arguments(t *testing.T) {
	should := require.New(t)
	argumentsDefPassed := ""
	_, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 0, `function hello(int $x, &$y, $z) {`, func(
			fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
			argumentsDefPassed = argumentsDef
			return fmt.Sprintf("%v();", functionName)
		})
	should.Equal(0, leftIdx)
	should.Equal(33, rightIdx)
	should.Equal("function hello(int $x, &$y, $z) { hello(); }\nfunction orig_hello(int $x, &$y, $z) {", substitution)
	should.Equal("int $x, &$y, $z", argumentsDefPassed)
}

func Test_newline_in_function_definition(t *testing.T) {
	should := require.New(t)
	_, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 0, "function\nhello() {", func(
			fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
			return fmt.Sprintf("%v();", functionName)
		})
	should.Equal(0, leftIdx)
	should.Equal(18, rightIdx)
	should.Equal("function\nhello() { hello(); }\nfunction\norig_hello() {", substitution)
}

func Test_line_no(t *testing.T) {
	should := require.New(t)
	tracepointLineNo := 0
	exitLineNo, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 15, "function\nhello() {", func(
			fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
			tracepointLineNo = lineNo
			return fmt.Sprintf("%v();", functionName)
		})
	should.Equal(15 + 1, exitLineNo)
	should.Equal(15 + 1, tracepointLineNo)
	should.Equal(0, leftIdx)
	should.Equal(18, rightIdx)
	should.Equal("function\nhello() { hello(); }\nfunction\norig_hello() {", substitution)
}

func Test_line_comment(t *testing.T) {
	should := require.New(t)
	_, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 0, `
##@@
##@@    public function getCityPrice($arrPriceList,$arrFilterCarType,$isFastCar,$open=1) {
##@@             $arrResult = array();
##@@    }

        /**
     * test
     */
    public function getFilterCityPrice($arrPriceList,$arrFilterCarType,$isFastCar,$open=1) {
		`, func(
			fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
			return fmt.Sprintf("%v();", functionName)
		})
	should.Equal(183, leftIdx)
	should.Equal(271, rightIdx)
	should.Equal("public function getFilterCityPrice($arrPriceList,$arrFilterCarType,$isFastCar,$open=1) { getFilterCityPrice(); }\npublic function orig_getFilterCityPrice($arrPriceList,$arrFilterCarType,$isFastCar,$open=1) {",
		substitution)
}

func Test_complex_function(t *testing.T) {
	should := require.New(t)
	_, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(
		"a.php", 0, `
    /**
     */
    public function getCurrentCityPriceByArea($sDistrict, $sCarLevel = '', $iProductId = GS_PRICE_GULFSTREAM, $iSchemaId = MILEAGE, $iComboType = 0, $iComboId = 0, $iAirport = 0, $sTimestamp = '') {
        if(empty($sDistrict)) {
		`, func(
			fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
			return fmt.Sprintf("%v();", functionName)
		})
	should.Equal(21, leftIdx)
	should.Equal(215, rightIdx)
	should.Equal("public function getCurrentCityPriceByArea($sDistrict, $sCarLevel = '', $iProductId = GS_PRICE_GULFSTREAM, $iSchemaId = MILEAGE, $iComboType = 0, $iComboId = 0, $iAirport = 0, $sTimestamp = '') { getCurrentCityPriceByArea(); }\npublic function orig_getCurrentCityPriceByArea($sDistrict, $sCarLevel = '', $iProductId = GS_PRICE_GULFSTREAM, $iSchemaId = MILEAGE, $iComboType = 0, $iComboId = 0, $iAirport = 0, $sTimestamp = '') {",
		substitution)
}


