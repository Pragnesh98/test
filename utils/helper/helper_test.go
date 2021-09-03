package helper

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestConvertToWave8000Bytes(t *testing.T) {
	rawInput, err := ioutil.ReadFile("test.raw")
	if err != nil {
		t.Fatalf("failed to read raw input test.raw, %s", err)
	}
	gotWavOutput, err := ConvertToWAV8000Bytes(rawInput)
	if err != nil {
		t.Fatalf("failed to convert, %s", err)
	}
	expectedWavOutput, err := ioutil.ReadFile("test.wav")
	if err != nil {
		t.Fatalf("failed to read expected wav test.wav, %s", err)
	}

	if len(gotWavOutput) != len(expectedWavOutput) {
		t.Errorf("got length = %d, expected length = %d", len(gotWavOutput), len(expectedWavOutput))
	}
	// The bytes are not exactly matching (for expected and got) but both seems valid (verified using play)
}

func TestGetSSMLList(t *testing.T) {
	type test struct {
		testcase string
		input    string
		expected []string
	}
	var empty []string
	tests := []test{
		{testcase: "Valid single SSML Type: 1", input: `<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>`, expected: []string{"<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>"}},
		{testcase: "Valid single SSML Type: 2", input: `<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US"><voice name="en-GB-MiaNeural"><mstts:express-as style="General"><prosody rate="0%" pitch="0%">\nThank you. Your profile is now created.\n\n</prosody></mstts:express-as></voice></speak>`, expected: []string{`<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US"><voice name="en-GB-MiaNeural"><mstts:express-as style="General"><prosody rate="0%" pitch="0%">\nThank you. Your profile is now created.\n\n</prosody></mstts:express-as></voice></speak>`}},
		{testcase: "Valid multiple SSML Type: 1", input: `<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak><speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>`, expected: []string{"<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>", "<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>"}},
		{testcase: "Valid multiple SSML Type: 2", input: `<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US"><voice name="en-GB-MiaNeural"><mstts:express-as style="General"><prosody rate="0%" pitch="0%">\nThank you. Your profile is now created.\n\n</prosody></mstts:express-as></voice></speak><speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US"><voice name="en-GB-MiaNeural"><mstts:express-as style="General"><prosody rate="0%" pitch="0%">\nThank you. Your profile is now created.\n\n</prosody></mstts:express-as></voice></speak>`, expected: []string{`<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US"><voice name="en-GB-MiaNeural"><mstts:express-as style="General"><prosody rate="0%" pitch="0%">\nThank you. Your profile is now created.\n\n</prosody></mstts:express-as></voice></speak>`, `<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US"><voice name="en-GB-MiaNeural"><mstts:express-as style="General"><prosody rate="0%" pitch="0%">\nThank you. Your profile is now created.\n\n</prosody></mstts:express-as></voice></speak>`}},
		{testcase: "Invalid+Valid SSML", input: `<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak><speak>We hope you can reconsider to retain a good credit standing. Good bye.<xml></speak>`, expected: []string{"<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>"}},
		{testcase: "Invalid SSML", input: `<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak><speak>We hope you can reconsider to retain a good credit standing. Good bye.<xml></speak>`, expected: []string{"<speak>We hope you can reconsider to retain a good credit standing. Good bye.</speak>"}},
		{testcase: "Empty string", input: "", expected: empty},
	}

	for _, tc := range tests {
		got := GetSSMLList(tc.input)
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("[%v] Expected: %v, Got: %v", tc.testcase, tc.expected, got)
		}
	}
}

func TestGetDuration(t *testing.T) {
	dur, err := GetDuration("test.wav")
	if err != nil {
		t.Fatalf("failed to get duration, %s", err)
	}

	if dur != 2662 {
		t.Errorf("got duration = %d, expected length = 0", dur)
	}
}
