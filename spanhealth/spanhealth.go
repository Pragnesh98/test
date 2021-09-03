package spanhealth

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/metrics"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// Span holds the details about span
var Span []*contracts.GetPipeHealth

// NewPipeHealth returns the PipeHealth struct
func NewPipeHealth() *contracts.GetPipeHealth {
	return &contracts.GetPipeHealth{}
}

// InitSpanHealth initializes the PipeHealthChecker
func InitSpanHealth() {
	// Run an infinite loop
	for {
		//Get all the SIP Spans
		sipSpans := getSIPSpanDetails()
		// Fill the SIP Spans with currently active channels
		fillSIPChannelCount(sipSpans)
		// Get all the PRI Spans
		priSpans := getPRISpanDetails()
		// Fill the PRI Spans with currently active channels
		fillPRIChannelCount(priSpans)
		// Merge PRI and SIP spans
		Span = append(sipSpans, priSpans...)
		// Print the output in the logger
		printOutput(Span)
		// Sleeping for the seconds specified in configuration
		time.Sleep(time.Duration(configmanager.ConfStore.PipeHealthDelay) * time.Second)
	}
}

func getSIPSpanDetails() []*contracts.GetPipeHealth {
	var sipSpans []*contracts.GetPipeHealth
	var sipIPs []string
	rCmdArguments := []string{"-rx", "sip show registry"}
	rOut, err := exec.Command("/usr/sbin/asterisk", rCmdArguments...).Output()
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Error while getting SIP Span Details. Error: [%#v]", err)
		return nil
	}
	rOutput := string(rOut[:])
	rperLineOut := strings.Split(rOutput, "\n")
	for _, rLine := range rperLineOut {
		sipSpan := NewPipeHealth()
		rPerLineFields := strings.Fields(rLine)
		if len(rPerLineFields) > 0 && len(rPerLineFields[0]) > 0 {
			sipIP := strings.Split(rPerLineFields[0], ":")[0]
			addr := net.ParseIP(sipIP)
			if addr == nil {
				continue
			}
			if inSlice(sipIPs, sipIP) {
				continue
			}
			sipIPs = append(sipIPs, sipIP)
			sipSpan.Name = sipIP
			if len(rPerLineFields) >= 4 && len(rPerLineFields[4]) > 0 && rPerLineFields[4] == "Registered" {
				sipSpan.Health.Up = true
			}
			sipSpans = append(sipSpans, sipSpan)
		}
	}

	pCmdArguments := []string{"-rx", "sip show peers"}
	pout, err := exec.Command("/usr/sbin/asterisk", pCmdArguments...).Output()
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Error while getting SIP Span Details. Error: [%#v]", err)
		return nil
	}
	pOutput := string(pout[:])
	pPerLineOut := strings.Split(pOutput, "\n")
	for _, line := range pPerLineOut {
		sipSpan := NewPipeHealth()
		perLineFields := strings.Fields(line)
		if len(perLineFields) > 0 && len(perLineFields[0]) > 0 {
			sipIP := strings.Split(perLineFields[0], "/")[0]
			addr := net.ParseIP(sipIP)
			if addr == nil {
				continue
			}
			if inSlice(sipIPs, sipIP) {
				continue
			}
			sipIPs = append(sipIPs, sipIP)
			sipSpan.Name = sipIP
			if len(perLineFields) >= 5 && len(perLineFields[5]) > 0 && perLineFields[5] == "OK" {
				sipSpan.Health.Up = true
			}
			sipSpans = append(sipSpans, sipSpan)
		}
	}
	totalSipSpan := NewPipeHealth()
	totalSipSpan.Name = "Total"
	totalSipSpan.Health.Up = true
	sipSpans = append(sipSpans, totalSipSpan)
	return sipSpans
}

func fillSIPChannelCount(sipSpans []*contracts.GetPipeHealth) {
	for _, sipSpan := range sipSpans {
		if sipSpan.Health.Up {
			sipSpan.Health.ChannelCount = getSIPChannelCount(sipSpan.Name)
		}
	}
	return
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func getSIPChannelCount(sipIP string) int {
	var cmd string
	cmd = "asterisk -rx \"core show channels count\" | grep " + sipIP + " | wc -l"
	if sipIP == "Total" {
		cmd = "asterisk -rx \"core show channels count\" | grep \"active call\" | cut -d\" \" -f1"
	}
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Failed to get the SIP channel count. SIPIP: [%s]. Error: [%#v]", sipIP, err)
		return 0
	}
	count, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Failed to convert string to integer. Error: [%#v]", err)
		return 0
	}
	return count
}

func getPRISpanDetails() []*contracts.GetPipeHealth {
	var priSpans []*contracts.GetPipeHealth
	cmdArguments := []string{"-rx", "pri show spans"}
	out, err := exec.Command("/usr/sbin/asterisk", cmdArguments...).Output()
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Error while getting PRI Span Details. Error: [%#v]", err)
		return nil
	}
	output := string(out[:])
	perLineOut := strings.Split(string(output), "\n")
	for _, line := range perLineOut {
		priSpan := NewPipeHealth()
		perLineFields := strings.Fields(line)
		if len(perLineFields) > 2 && len(perLineFields[2]) > 0 {
			priSpanNum := strings.Split(perLineFields[2], "/")[0]
			priSpan.Name = priSpanNum
			if len(perLineFields) >= 3 && len(perLineFields[3]) > 0 && strings.HasPrefix(perLineFields[3], "Up") {
				priSpan.Health.Up = true
			}
			priSpans = append(priSpans, priSpan)
		}
	}
	return priSpans
}

func fillPRIChannelCount(priSpans []*contracts.GetPipeHealth) {
	for _, priSpan := range priSpans {
		if priSpan.Health.Up {
			priSpan.Health.ChannelCount = getPRIChannelCount(priSpan.Name)
		}
	}
}

func getPRIChannelCount(priSpanNum string) int {
	cmdArguments := []string{"-rx", "pri show span " + priSpanNum}
	out, err := exec.Command("/usr/sbin/asterisk", cmdArguments...).Output()
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Error while getting PRI Channel Count. PRISpan: [%s]. Error: [%#v]", priSpanNum, err)
		return 0
	}
	output := string(out[:])
	perLineOut := strings.Split(output, "\n")
	for _, line := range perLineOut {
		perLineFields := strings.Fields(line)
		if len(perLineFields) > 1 && len(perLineFields[0]) > 0 && perLineFields[0] == "Total" && len(perLineFields[1]) > 0 {
			splitCalls := strings.Split(perLineFields[1], ":")
			if len(splitCalls) >= 2 {
				count, err := strconv.Atoi(splitCalls[1])
				if err != nil {
					ymlogger.LogErrorf("PipeHealth", "Error while converting string to integer. Error: [%#v]", err)
					return 0
				}
				return count
			}
		}
	}
	return 0
}

func printOutput(spanDetails []*contracts.GetPipeHealth) {
	var output string
	for _, spanDetail := range spanDetails {
		str := fmt.Sprintf("\"%s\":[%v, %d], ", spanDetail.Name, spanDetail.Health.Up, spanDetail.Health.ChannelCount)
		output = output + str
		sendMetric(spanDetail)
	}
	ymlogger.LogInfof("PipeHealth", "[%s]", output)
}

func sendMetric(spanDetail *contracts.GetPipeHealth) {
	// Send event to New Relic
	eventData := map[string]interface{}{
		"span_name":  spanDetail.Name,
		"channel_up": strconv.FormatBool(spanDetail.Health.Up),
		"value":      spanDetail.Health.ChannelCount,
	}
	if err := newrelic.SendCustomEvent("call_pipehealth", eventData); err != nil {
		ymlogger.LogErrorf("PipeHealth", "Failed to send metric to new relic. Error: [%#v]", err)
	}

	filters := make(map[string]string)
	fields := make(map[string]interface{})
	filters["span_name"] = spanDetail.Name
	filters["channel_up"] = strconv.FormatBool(spanDetail.Health.Up)
	fields["value"] = spanDetail.Health.ChannelCount
	metric, err := metrics.NewMetric("call.pipehealth", filters, fields)
	if err != nil {
		ymlogger.LogErrorf("PipeHealth", "Failed to create metric. Error: [%#v]", err)
		return
	}
	if err := metrics.SendMetric(metric); err != nil {
		ymlogger.LogErrorf("PipeHealth", "Failed to send metrics. Error: [%#v]", err)
	}
	return
}
