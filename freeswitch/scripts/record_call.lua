--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  record_call.lua --> call recording on demand
--]]
        conference_name = argv[1];
        uuid = argv[2];

        script_dir = freeswitch.getGlobalVariable("script_dir")
        dofile(script_dir.."/logger.lua");
        if(script_dir == "/usr/local/freeswitch/scripts") then
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/local/share/lua/5.1/?.lua;" .. package.path
        else
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        end

--general functions
        require "resources.functions.base64";
        require "resources.functions.trim";
        require "resources.functions.file_exists";
        require "resources.functions.explode";
        require "resources.functions.format_seconds";
        require "resources.functions.mkdir";

--      package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
--      package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        xml = require "xml";

--prepare the api object
        api = freeswitch.API();

--check if the conference exists
        cmd = "conference "..conference_name.." xml_list";
        freeswitch.consoleLog("INFO","" .. cmd .. "\n");
        result = api:executeString(cmd)
        if string.match(result, "not found") then
                conference_exists = false;
        else
               conference_exists = true;
        end

confernce_record = session:getVariable("conference_record")
--freeswitch.consoleLog("INFO","conference_recording : "..tostring(confernce_record).."\n")
--if(tostring(conference_record) == 'true') then
--freeswitch.consoleLog("INFO","conference_recording : "..tostring(confernce_record).."\n")
     if (conference_exists) then
	result = string.match(result,[[<conference (.-)>]],1);
	conference_session_uuid = string.match(result,[[uuid="(.-)"]],1);
	freeswitch.consoleLog("INFO","[call_record] conference_session_uuid: " .. conference_session_uuid .. "\n");
	conf_recording = session:getVariable("conf_recording")
	session:setVariable("leader_enable_recording", "true");
	
	if(conf_recording == "true") then
		sounds_dir = session:getVariable("prompt_dir");
		audio_file = sounds_dir.."/70.mp3"
		Logger.info("[MindBridge] audio_file : " .. tostring(audio_file) .. "");
		confirm = session:playAndGetDigits(1, 1, 1,10000, "#", audio_file, "", "");
		if(confirm == "1") then
	               session:setVariable("conf_call_recording", "stoped");
        	       session:execute("unset", "conf_recording");
			
			conference_call_recording = session:getVariable("conference_call_recording");
			cmd = "conference "..conference_name.." norecord "..tostring(conference_call_recording);
			freeswitch.consoleLog("notice", "[call_record] stop call record api cmd: " .. cmd .. "\n");
			response = api:executeString(cmd);
			session:streamFile(sounds_dir.."/43.mp3");
			return 0;
		elseif (confirm == "*") then
			Logger.info("[MindBridge] cancelled reuest");
			return 0;
		end

--		session:streamFile(sounds_dir.."/43.mp3");
--		conference_call_recording = session:getVariable("conference_call_recording");
--		cmd = "conference "..conference_name.." norecord "..tostring(conference_call_recording);
--		freeswitch.consoleLog("notice", "[call_record] stop call record api cmd: " .. cmd .. "\n");
--		response = api:executeString(cmd);
	else
		MaxTry = 3;
		Try = 0;
		if (wait_mod == "true" and member_type ~= "moderator") then
                                Logger.notice("[MindBridge] : Call will not be recording");
                                --don't start recording yet
                        else
                                Logger.notice("[MindBridge] : Call will be recording");
                                recordings_dir = "/usr/local/freeswitch/recordings";
                                                recordings_dir = recordings_dir.."/archive/"..os.date("%Y", start_epoch).."/"..os.date("%b", start_epoch).."/"..os.date("%d", start_epoch);
                                                mkdir(recordings_dir);
                                                recording = recordings_dir.."/"..conference_name.."_"..conference_session_uuid..".mp3";


                                                session:setVariable("conf_recording","true");
                                                sounds_dir = session:getVariable("prompt_dir");
                                                --session:streamFile(sounds_dir.."/37.mp3");
                                                session:streamFile(sounds_dir.."/start_recording.mp3");
                                                cmd = "conference "..conference_name.." record "..recording;
                                                freeswitch.consoleLog("notice", "[call_record] : start call record api cmd: " .. cmd .. "\n");
                                                response = api:executeString(cmd);


                        end


--[[		while(Try < MaxTry) do
			Try = Try + 1;
			min_digits = 4;
			max_digits = 20;
			max_tries = 1;
			digit_timeout = 20000;
			flags = "";

			sounds_dir = session:getVariable("prompt_dir");
			prompt_audio_file = sounds_dir.."/36.mp3"
			Logger.info("[MindBridge] Playing file name : " .. tostring(prompt_audio_file) .. "");
			file_number = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", prompt_audio_file, "", "");
			
			if(file_number == "*") then
				freeswitch.consoleLog("notice", "[call_record] : call recording request cancelled\n");
			else
				freeswitch.consoleLog("notice", "[call_record] : "..tostring(file_number).."\n");
				if(file_number == "" or file_number == nil or file_number == "nil") then
					freeswitch.consoleLog("notice", "[call_record] : file verification failed \n");
					file_verified = false
					Try = Try + MaxTry + 1
					return 0;
				else
					audio_file = sounds_dir.."/38.mp3"
					session:streamFile(audio_file);
					session:sleep(500);

					session:execute("say", "en number iterated  "..file_number);
					file_verified = true
				end

				start_epoch = os.time();
				if(file_verified) then
					prompt_audio_file = sounds_dir.."/start_recording.mp3"
					Logger.info("[MindBridge] Playing file name : " .. tostring(prompt_audio_file) .. "");
					confirm = session:playAndGetDigits(1, 1, max_tries, digit_timeout, "#", prompt_audio_file, "", "");

					if(confirm == "1") then
						recordings_dir = "/usr/local/freeswitch/recordings";
						recordings_dir = recordings_dir.."/archive/"..os.date("%Y", start_epoch).."/"..os.date("%b", start_epoch).."/"..os.date("%d", start_epoch);
						mkdir(recordings_dir);
						recording = recordings_dir.."/"..file_number.."_"..conference_session_uuid..".wav";

						--conf_call_recording = session:getVariable("conf_call_recording");
						--session:setVariable("conf_call_recording", "started");
						--cmd = "uuid_setvar "..tostring(uuid).." conference_call_recording "..recording;
						--freeswitch.consoleLog("notice", "[call_record] cmd: " .. cmd .. "\n");
						--response = api:executeString(cmd);

						session:setVariable("conf_recording","true");
						sounds_dir = session:getVariable("prompt_dir");
						--session:streamFile(sounds_dir.."/37.mp3");
						--session:streamFile(sounds_dir.."/start_recording.mp3");
						cmd = "conference "..conference_name.." record "..recording;
						freeswitch.consoleLog("notice", "[call_record] : start call record api cmd: " .. cmd .. "\n");
						response = api:executeString(cmd);
						Try = Try + MaxTry + 1
						return 0;
					elseif (confirm == "2") then
						Logger.info("[MindBridge] Re-enter conference file number");
						Try = 1
					elseif (confirm == "*") then
						file_verified = false
						Try = Try + MaxTry + 1
						Logger.info("[MindBridge] cancelled reuest");
						return 0;
					end
				end
			end
		end--]]
	end
   end

--end
