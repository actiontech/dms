// Package dlp sdkdeidentify.go implements deidentify related APIs
package dlp

import (
	"encoding/json"
	"fmt"
	"github.com/bytedance/godlp/dlpheader"
	"github.com/bytedance/godlp/errlist"
)

// public func
// Deidentify detects string firstly, then return masked string and results
// 对string先识别，然后按规则进行打码
func (I *Engine) Deidentify(inputText string) (outputText string, retResults []*dlpheader.DetectResult, retErr error) {
	defer I.recoveryImpl()
	if !I.hasConfiged() { // not configed
		panic(errlist.ERR_HAS_NOT_CONFIGED)
	}
	if I.hasClosed() {
		return "", nil, errlist.ERR_PROCESS_AFTER_CLOSE
	}
	if I.isOnlyForLog() {
		return inputText, nil, errlist.ERR_ONLY_FOR_LOG
	}
	if len(inputText) > DEF_MAX_INPUT {
		return inputText, nil, fmt.Errorf("DEF_MAX_INPUT: %d , %w", DEF_MAX_INPUT, errlist.ERR_MAX_INPUT_LIMIT)
	}
	outputText, retResults, retErr = I.deidentifyImpl(inputText)
	return
}

// DeidentifyMap detects KV map firstly,then return masked map
// 对map[string]string先识别，然后按规则进行打码
func (I *Engine) DeidentifyMap(inputMap map[string]string) (outMap map[string]string, retResults []*dlpheader.DetectResult, retErr error) {
	defer I.recoveryImpl()

	if !I.hasConfiged() { // not configed
		panic(errlist.ERR_HAS_NOT_CONFIGED)
	}
	if I.hasClosed() {
		return nil, nil, errlist.ERR_PROCESS_AFTER_CLOSE
	}
	if len(inputMap) > DEF_MAX_ITEM {
		return inputMap, nil, fmt.Errorf("DEF_MAX_ITEM: %d , %w", DEF_MAX_ITEM, errlist.ERR_MAX_INPUT_LIMIT)
	}
	outMap, retResults, retErr = I.deidentifyMapImpl(inputMap)
	return
}

// DeidentifyJSON detects JSON firstly, then return masked json object in string formate and results
// 对jsonText先识别，然后按规则进行打码，返回打码后的JSON string
func (I *Engine) DeidentifyJSON(jsonText string) (outStr string, retResults []*dlpheader.DetectResult, retErr error) {
	defer I.recoveryImpl()

	if !I.hasConfiged() { // not configed
		panic(errlist.ERR_HAS_NOT_CONFIGED)
	}
	if I.hasClosed() {
		return jsonText, nil, errlist.ERR_PROCESS_AFTER_CLOSE
	}
	outStr = jsonText
	if results, kvMap, err := I.detectJSONImpl(jsonText); err == nil {
		retResults = results
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(jsonText), &jsonObj); err == nil {
			//kvMap := I.resultsToMap(results)
			outObj := I.dfsJSON("", &jsonObj, kvMap, true)
			if outJSON, err := json.Marshal(outObj); err == nil {
				outStr = string(outJSON)
			} else {
				retErr = err
			}
		} else {
			retErr = err
		}
	} else {
		retErr = err
	}
	return
}

// private func
// deidentifyImpl implements Deidentify string
func (I *Engine) deidentifyImpl(inputText string) (outputText string, retResults []*dlpheader.DetectResult, retErr error) {
	outputText = inputText // default same text
	if arr, err := I.detectImpl(inputText); err == nil {
		retResults = arr
		if out, err := I.deidentifyByResult(inputText, retResults); err == nil {
			outputText = out
		} else {
			retErr = err
		}
	} else {
		retErr = err
	}
	return
}

// deidentifyMapImpl implements DeidentifyMap
func (I *Engine) deidentifyMapImpl(inputMap map[string]string) (outMap map[string]string, retResults []*dlpheader.DetectResult, retErr error) {
	outMap = make(map[string]string)
	if results, err := I.detectMapImpl(inputMap); err == nil {
		if len(results) == 0 { // detect nothing
			return inputMap, results, nil
		} else {
			outMap = inputMap
			for _, item := range results {
				if orig, ok := outMap[item.Key]; ok {
					if out, err := I.deidentifyByResult(orig, []*dlpheader.DetectResult{item}); err == nil {
						outMap[item.Key] = out
					}
				}
			}
			retResults = results
		}
	} else {
		outMap = inputMap
		retErr = err
	}
	return
}

// deidentifyByResult concatenate MaskText
func (I *Engine) deidentifyByResult(in string, arr []*dlpheader.DetectResult) (string, error) {
	out := make([]byte, 0, len(in)+8)
	pos := 0
	inArr := S2B(in)
	for _, res := range arr {
		if pos < res.ByteStart {
			out = append(out, []byte(inArr[pos:res.ByteStart])...)
		}
		out = append(out, []byte(res.MaskText)...)
		pos = res.ByteEnd
	}
	if pos < len(in) {
		out = append(out, []byte(inArr[pos:])...)
	}
	outStr := B2S(out)
	return outStr, nil
}

// resultsToMap convert results array into Map[Key]=MaskText
func (I *Engine) resultsToMap(results []*dlpheader.DetectResult) map[string]string {
	kvMap := make(map[string]string)
	for _, item := range results {
		kvMap[item.Key] = item.MaskText
	}
	return kvMap
}
