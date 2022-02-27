--[[
The Initial Developer of the Original Code is

Portions created by the Initial Developer are Copyright (C)
the Initial Developer. All Rights Reserved.

Contributor(s):

dialout.lua --> conference controls
--]]

script_dir = freeswitch.getGlobalVariable("script_dir")
dofile(script_dir.."/logger.lua");
if(script_dir == "/usr/local/freeswitch/scripts") then
		package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
		package.path = "/usr/local/share/lua/5.1/?.lua;" .. package.path
else
		package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
		package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
end

api = freeswitch.API();
xml = require "xml";

local function Dialout(session, conference_name, CallUUID)
--sounds_dir = session:getVariable("prompt_dir");
sounds_dir =freeswitch.getGlobalVariable("prompt_dir")

--conference_name = session:getVariable("conference_name");
conference_name = argv[1]
 CallUUID = argv[2]
        Logger.notice ("[ConfAPI] : conference_name : " ..tostring(conference_name).." CallUUID :" ..tostring(CallUUID).." \n")
--        MemberID, speak_a, hear_a = getParticipantMuteStatus(CallUUID, conference_name);

--        allUUIDs = all_participant_uuids(conference_name)

--CallUUID = session:getVariable("uuid");
Logger.notice ("[ConfAPI] : conference dialout request")

--session:setVariable("proceed_ivr", "true");
cmd = "conference " .. conference_name .. " moh off"
response = api:executeString(cmd);
freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));


Retry = 0;
MaxReTry = 1;
while(Retry < MaxReTry ) do
		if (session:ready()) then
				min_digits = 1;
				max_digits = 20;
				max_tries = 1;
				digit_timeout = 5000;

				prompt_audio_file = sounds_dir.."/host_dialout.wav"
				Logger.info("[MindBridge] enter your mobile number : " .. tostring(prompt_audio_file) .. "");
				number = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", prompt_audio_file, "", "");
                                Logger.info("[MindBridge] : USER INPUT :: "..tostring(number));
				if(number == "*") then
					session:streamFile(sounds_dir.."/66.mp3"); ---operator req canceled
                                         Logger.info("[MindBridge] : THANK YOU!!!");
                                          Retry = MaxReTry + 1;
                                           break;

				else
--                      number = number:gsub("%p","")

				Retry = Retry + 1;
				if(number == "" or number == nil or number == "nil") then
						session:streamFile(sounds_dir.."/53.mp3"); ---invalid digit re-enter
				else
						session:streamFile(sounds_dir.."/48.mp3"); --- phone no entered is
						session:set_tts_params("flite", "kal");
						number = number:gsub("^%s+", ""):gsub("%s+$", "")
						if number ~= nil then
								--session:speak(number)
								Logger.info("[MindBridge] : You entered number : " .. tostring(number) .. "");
								session:say(number, "en", "name_spelled", "iterated");
						end

				Retry1 = 0;
				while(Retry1 < MaxReTry ) do

						min_digits = 1;
						max_digits = 1;
						max_tries = 1;
						digit_timeout = 10000;

						announce_sound = sounds_dir.."/49.mp3" ---press 1 to call 2 to re-enter.....
						Logger.info("[MindBridge] confirm to dialout : " .. tostring(announce_sound) .. "");
						user_input = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", announce_sound, "","^(.*)$");
						Logger.info("[MindBridge] : User selected option : " .. tostring(user_input) .. "");

						if(user_input == "1") then
							session:execute("playback", sounds_dir.."/option_not_available.wav");
							--[[	Retry = MaxReTry + 1;
								Retry1 = MaxReTry + 1;
								sofia_str = "{origination_caller_id_number='+912238013500',origination_caller_id_name='+912238013500'}sofia/internal/"..tostring(number).."@10.8.48.190:5060;fs_path=sip:10.17.109.157"
								cmd = "conference "..tostring(conference_name).." dial "..sofia_str.." Conference Conference"
								response = api:executeString(cmd);]]--
								break;
						elseif (user_input == "2") then
								Logger.info("[MindBridge] : Retry Again!!!");
								Retry = 0;
								Retry1 = MaxReTry + 1;
						elseif (user_input ==  "*") then
								Logger.info("[MindBridge] : THANK YOU!!!");
								Retry = MaxReTry + 1;
								Retry1 = MaxReTry + 1;
								break;
						else
								Logger.info("[MindBridge] : NO INPUT!!!");
								Retry1 = Retry1 + 1;
						end
				end

				end
				end
		end
end

		cnt = 0
		session:sleep(1000);
		Keypad = require 'keypad_commands-loop'
		Keypad(session, conference_name, aleg_uuid, cnt)
end

return Dialout
