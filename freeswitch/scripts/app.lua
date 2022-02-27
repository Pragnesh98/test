--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  app.lua --> conference centers
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

--general functions
        require "resources.functions.base64";
        require "resources.functions.trim";
        require "resources.functions.file_exists";
        require "resources.functions.explode";
        require "resources.functions.format_seconds";
        require "resources.functions.mkdir";
        local json = require "resources.functions.lunajson"

--      package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
--      package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        xml = require "xml";

        debug["debug"] = false;
--set variables
        flags = "";
        max_tries = 3;
        digit_timeout = 5000;

--prepare the api object
        api = freeswitch.API();

        temp_dir = "/tmp/"
        flags = ""

        UrL = freeswitch.getGlobalVariable("UrL")
        uuid = session:getVariable("uuid");
        session:answer();

        function create_conference_api(conference_server_ip, conference_name)
                conf_id = session:getVariable("conf_id");
                payload = '{"conference_id":"'..tostring(conf_id)..'","conference_server_ip":"'..tostring(conference_server_ip)..'","conference_name":"'..tostring(conference_name)..'"}'
                APIcmd = "curl --location --request PUT '"..tostring(UrL).."conference-server-ip' -H 'content-type: application/json'  -d '"..tostring(payload).."'"
                Logger.info ("[MindBridge] : Conf Server POST Request : "..tostring(APIcmd).."");
                curl_response = api:execute("system", APIcmd);
                Logger.debug ("[MindBridge] : Conf Server POST Response : "..tostring(curl_response).."");
        end

        function delete_conference_api(conference_server_ip, conference_name)
                APIcmd = "curl --location --request DELETE '"..tostring(UrL).."delete-conference-server-ip/"..tostring(conference_name).."'"
                Logger.info ("[MindBridge] : Conf Server DELETE Request : "..tostring(APIcmd).."");
                curl_response = api:execute("system", APIcmd);
                Logger.debug ("[MindBridge] : Conf Server DELETE Response : "..tostring(curl_response).."");
                Logger.notice ("[MindBridge] : Delete Conference From DB");
        end

        function get_pin_number(prompt_audio_file, wrong_pin_music, after_pin_music)
                if (session:ready()) then
                        min_digits = 2;
                        max_digits = 20;
                        max_tries = 1;
                        digit_timeout = 5000;
                        flags = "";

			 if(conference_ID ~= nil) then
                	        conference_name = session:getVariable("sip_h_X-conference_name");
                        	pin_number = session:getVariable("sip_h_X-PIN");
	                        solution_type = session:getVariable("sip_h_X-solution");
        	        else

				Logger.info("[MindBridge] Playing file name : " .. tostring(prompt_audio_file) .. "");
                        	pin_number = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", prompt_audio_file, "", "\\d+");
                        	--remove non numerics
                        	pin_number = pin_number:gsub("%p","")
			end
                               Logger.info ("[MindBridge] : Conf PIN number : "..tostring(pin_number).."\n");

                        solution_type = session:getVariable("solution_type");
                        if (tostring(pin_number) ~= "") then
                                APIcmd = "curl --location --request GET '"..tostring(UrL).."did/verify-pin/"..tostring(destination_number).."/"..tostring(pin_number).."'"
                                Logger.info ("[MindBridge] : Conf PIN verification APIcmd : "..tostring(APIcmd).."");
                                curl_response = api:execute("system", APIcmd);
                                Logger.debug ("[MindBridge] : Conf PIN verification API Response : "..tostring(curl_response).."");

                                local encode = json.decode (curl_response)
                                local MessageRes = encode['message']
                                local Msg = encode['Msg']
                                if(MessageRes == "Not Found" or Msg == "Failed") then
                                        if(wrong_pin_music) then
                                                session:streamFile(wrong_pin_music);
					end
					if pin_number ~= nil then
						session:execute("playback",sounds_dir.."/invalid_pin.wav")
                                        	pin_number = "not found";
						chairperson_pin, conference_id, member_type, flags = nil,nil,nil,nil;
					else
						pin_number = nil;
					end
                                else
					session:setVariable("pin_number", tostring(pin_number))
					session:setVariable("conference_code", pin_number);
                                        conf_id = encode['did_conference_map']['conference_id']
                                        session:setVariable("conf_id", tostring(conf_id));
                                        --pin_number = "verified";
                                        conf_pin_number = encode['did_conference_map']['pin_number']
                                        chairperson_pin = encode['did_conference_map']['chairperson_pin']
                                        active_conference_server_ip = encode['did_conference_map']['conference_server_ip']
					member_type = encode['did_conference_map']['pin_type']
					
					if(solution_type == "Reservation Less" and member_type == "CHAIRPERSON") then
                                        	pin_number = nil;
                                		return nil, nil, nil, nil,nil;
					end

                                        announce_name = encode['did_conference_map']['announce_name']
                                        max_members = encode['did_conference_map']['max_members']
                                        service_id = encode['did_conference_map']['service_id']
                                        session:setVariable("service_id", tostring(service_id));

                                        conference_name = encode['did_conference_map']['conference_name']
                                        conference_record  = encode['did_conference_map']['conference_recording']
                                        conference_wait_for_moderator  = encode['did_conference_map']['conference_wait_for_moderator']
                                        conference_mute  = encode['did_conference_map']['conference_mute ']
					conference_continuation  = encode['did_conference_map']['conference_continuation ']
					participant_mute  = encode['did_conference_map']['participant_mute']
                                        Logger.debug ("[MindBridge] : participant_mute : " ..tostring(participant_mute));
					session:setVariable("participant_mute", tostring(participant_mute));
					party_lock_unlock  = encode['did_conference_map']['party_lock_unlock']
                                        Logger.debug ("[MindBridge] : party_lock_unlock : " ..tostring(party_lock_unlock));
					session:setVariable("party_lock_unlock", tostring(party_lock_unlock));
					enable_disable  = encode['did_conference_map']['enable_disable']
                                        Logger.debug ("[MindBridge] : enable_disable : " ..tostring(enable_disable));
					session:setVariable("enable_disable", tostring(enable_disable));
					--[[if(conference_continuation == "" or conference_continuation == nil or conference_continuation == "nil") then
						conference_continuation = "allow"
					end]]--
                                        start_time  = encode['did_conference_map']['start_time ']
                                        end_time  = encode['did_conference_map']['end_time ']
                                        Logger.debug ("[MindBridge] : start_time : " ..tostring(start_time));
                                        Logger.debug ("[MindBridge] : end_time : " ..tostring(end_time));
				
					entry_announce  = encode['did_conference_map']['entry_announce ']
					exit_announce  = encode['did_conference_map']['exit_announce ']
					if(exit_announce == "" or exit_announce == nil or exit_announce == "nil") then
						conference_exit_sound = "tone_stream://v=-20;%(90,60,620);/%(90,60,440)";
					else
						conference_exit_sound = exit_announce
					end
					session:setVariable("exit_sound", conference_exit_sound);

					if(entry_announce == "" or entry_announce == nil or entry_announce == "nil") then
						conference_enter_sound = "tone_stream://v=-20;%(100,1000,100);v=-20;%(90,60,440);%(90,60,620)";
					else
						conference_enter_sound = entry_announce
					end
					session:setVariable("enter_sound", conference_enter_sound);
					Logger.debug ("[MindBridge] : conference_enter_sound : " ..tostring(conference_enter_sound));
					Logger.debug ("[MindBridge] : conference_exit_sound : " ..tostring(conference_exit_sound));

                                        if(member_type == "CHAIRPERSON") then
                                                member_type = "moderator";
                                        else
                                                member_type = "participant"
                                        end

                                        if(conference_wait_for_moderator == "true") then
                                             if (member_type == "participant") then
                                                     flags = flags .. "wait-mod";
                                             end
                                                session:setVariable("wait_mod", tostring(conference_wait_for_moderator));
                                        end
                                        if(conference_mute == "true") then
                                                if (member_type == "participant") then
                                                        flags = flags .. "|mute";
                                                end
                                        end
                                        if (member_type == "moderator") then
                                                flags = flags .. "|moderator|endconf";
						Identity_name = "Host Port 1-1-1-2 97";
						session:setVariable("Identity_name:", tostring(Identity_name));
                                        	Logger.debug ("[MindBridge] : flags : " ..tostring(Identity_name));
					else
						Identity_name = "Guest Port 1-1-1-2 97";
						session:setVariable("Identity_name","Guest Port 1-1-1-4 2");
                                        	Logger.debug ("[MindBridge] : flags : " ..tostring(Identity_name));
                                        end
                                        if(flags == nil or flags == "" or flags == "nil") then
                                                flags = "";
                                        end
                                        Logger.debug ("[MindBridge] : flags : " ..tostring(flags));
					session:setVariable("conference_continuation", conference_continuation);
                                        session:setVariable("conference_permanent_wait_mod_moh", "false");
                                        session:setVariable("conference_record", tostring(conference_record));
                                        session:setVariable("member_type", tostring(member_type));
                                        session:setVariable("announce_name", tostring(announce_name));
                                        session:setVariable("active_conference_server_ip", tostring(active_conference_server_ip));
					conference_server_ip = freeswitch.getGlobalVariable("conference_server_ip")
                                        Logger.debug ("[MindBridge] : conference_server_ip : " ..tostring(conference_server_ip).." activer_conference_ip: "..tostring(active_conference_server_ip).."\n");
					
					if((active_conference_server_ip == conference_server_ip) or (active_conference_server_ip == nil or active_conference_server_ip == "" or active_conference_server_ip == "nil")) then
						Logger.notice("[MindBridge] : Conference Call");
					else
                                                Logger.notice("[MindBridge] : Conference ID ["..tostring(conference_name).."] Active On Server ["..tostring(active_conference_server_ip).."]");
                                                destination_number = session:getVariable("destination_number");
                                        	session:setVariable("hangup_after_bridge", "true");
                                                session:execute("bridge","{hangup_after_bridge=true,sip_h_X-PIN="..tostring(pin_number)..",sip_h_X-solution="..tostring(solution_type)..",sip_h_X-conference_name="..tostring(conference_name)..",sip_h_X-conference-ID="..tostring(conf_data).."}sofia/external/"..tostring(destination_number).."@"..tostring(active_conference_server_ip));
                                                session:hangup();
                                                return 0;						 
					end

					if(conference_name == "" or conference_name == nil or conference_name == "nil") then
						session:setVariable("conference_started", "false");
						local conference_server_ip = freeswitch.getGlobalVariable("conference_server_ip")

                                                uuid = session:get_uuid()
						create_uuid = api:executeString("create_uuid");
                                                did_number = session:getVariable("sip_to_user");
                                                did_number = string.gsub(did_number,"^+","");
                                                conference_id = ""..tostring(conf_id)..""..tostring(conf_pin_number)..""..tostring(chairperson_pin).."_mindbridge"
                                                Logger.debug ("[MindBridge] : Conf Name : " ..tostring(conference_id));
                                                session:setVariable("conference_status", "false");
						create_conference_api(conference_server_ip, conference_id)
                                        else
						session:setVariable("conference_started", "true");
                                                conference_id = conference_name
                                        end
                                        session:setVariable("conference_name", conference_id);
                                        session:setVariable("max_members", tostring(max_members));

                                        Logger.debug ("[MindBridge] : active conference server ip : " ..tostring(conference_server_ip));
                                        Logger.debug ("[MindBridge] : chairperson_pin : " ..tostring(chairperson_pin));
                                        Logger.debug ("[MindBridge] : pin_type : " ..tostring(member_type));
                                        Logger.debug ("[MindBridge] : conference_record : " ..tostring(conference_record));
                                        Logger.debug ("[MindBridge] : max_members : " ..tostring(max_members));
                                        Logger.debug ("[MindBridge] : Conf PIN verified Successfully");
                                        pin_number = "verified";
                                        if(after_pin_music) then
                                                session:streamFile(after_pin_music);
                                        else
                                                --session:streamFile(sounds_dir.."/thank_you.wav");
                                        end
                                        session:sleep(500);
                                end
                        else
                                pin_number = nil;
                                conference_id = nil;
                        end

                        if (pin_number == nil) then
                                return nil, nil, nil, nil,nil;
                        else
                                Logger.info("[MindBridge] Conf PIN : " .. tostring(pin_number) .. "");
                                return pin_number, chairperson_pin, conference_id, member_type, flags;
                        end
                else
                        session:hangup();
                        return 0;
                end
        end

        function ModeratoriSOn(unique_id,conf_name)
		cnt = 0
                id, hear_value, speak_value = nil, nil, nil;
                is_moderator, is_moderator1 = false,false;
                isLock = false;
                result = api:executeString("conference "..conf_name .. " xml_list")
                if string.match(result, "not found") then
                        Logger.notice("[MindBridge] : No Active conference "..tostring(conf_name))
                        result = nil;
                        return 0,is_moderator, false;
                end
                if(result == nil or result == "nil" or result == "") then
                        return 0,false, false;
                end
                result = result:gsub("<variables></variables>", "")

                xmltable = xml.parse(result)
                for i=1,#xmltable do
                        if xmltable[i].tag then
                                local tag = xmltable[i]
                                local subtag = tag[i]

                                isLock = tag.attr.locked
                                if subtag.tag == "members" then
                                        if (debug["debug"]) then
                                                Logger.notice("[MindBridge] : subtag : ".. tostring(subtag) .. "")
                                        end
                                        for j = 1,#subtag do
                                                for k = 1,#subtag[j] do
                                                        if (debug["debug"]) then
                                                                Logger.notice("[MindBridge] : subtag[j][k] : ".. tostring(subtag[j][k]) .. "")
                                                        end
                                                        if( subtag[j][k] ~= nil) then
                                                                if tostring(subtag[j][k].tag) == 'uuid' then
                                                                        if  subtag[j][k][1] == unique_id then
                                                                                if (debug["debug"]) then
                                                                                        Logger.notice("[MindBridge] : Log  Data "..unique_id.."\tConference_Name : "..conf_name.."\nFS RESPONSE :  "..result.."");
                                                                                end
                                                                        end
                                                                end

                                                                s1 = subtag[j]
                                                                for i, j in ipairs(s1) do

                                                                              --Logger.notice("[MindBridge] : cnt : "..tostring(cnt));
								--if cnt == 0 then

                                                                        if( string.find(tostring(j),"<id>") ~= nil )then
                                                                                id = tostring(j):match("id>(.-)</id>")
											id1 = id
--                                                                              Logger.notice("[MindBridge] : CONFID : "..tostring(id));
                                                                        elseif(string.find(tostring(j),"<is_moderator>") ~= nil ) then
										
                                                                                is_moderator = tostring(j):match("is_moderator>(.-)</is_moderator>")
                                                                              	if is_moderator == "true" then
											is_moderator1="true";
										end
--                                                                              Logger.notice("[MindBridge] : IS_MODERATOR : "..tostring(is_moderator));
									end
                                                                --end
								end
								cnt = cnt+1;
                                                        end
                                                end
                                        end
                                end
                        end
                        i = i+1;
                end
		
		if is_moderator1 == "true" then
			is_moderator = is_moderator1;
		end

                Logger.notice("[MindBridge] : ID : "..tostring(id).." IS_MODERATOR : ["..tostring(is_moderator).."]");
                return id, is_moderator, isLock
        end
-- 		function onInput(s, type, obj)
-- 			if (type == "dtmf" and obj['digit'] == '#') then
-- 				return "break";
-- 			end
-- 		end


        function members_counts(conference_name)
                cmd = "conference "..tostring(conference_name).." list count";
                member_count = api:executeString(cmd);
                if string.match(member_count, "not found") then
                        member_count = "0";
                end
                return member_count
        end

        function record_your_name(record_name_prompt)
                --prompt for the name of the caller
                session:execute("playback", record_name_prompt);
            --    session:execute("playback", "tone_stream://v=-5;%%(500,0,500.0)");beep.wav
	        session:execute("playback", "/usr/local/freeswitch/prompt/joing_beep.wav");
	--	uuid = session:get_uuid();
          --      cmd = "uuid_audio "..uuid.." start read level 3";
	--	Logger.notice("[MindBridge] : cmd :" ..tostring(cmd).."\n");
                response = api:executeString(cmd);
		--record the response
                max_len_seconds = 2;
                silence_threshold = 30;
                silence_secs = 5;
                session:recordFile(temp_dir:gsub("\\","/") .. "/conference-"..uuid..".mp3", max_len_seconds, silence_threshold, silence_secs);
		--user_number = session:playAndGetDigits(1, 1, 1,4000, "#", "", "", "");
		session:streamFile(sounds_dir.."/thank_you.wav");
        end

        function announce_your_name(conference_name, joined_prompt)
		--[[	cmd = "conference "..conference_name.." moh off";
			freeswitch.consoleLog("notice", "[moh] cmd: " .. cmd .. "\n");
			response = api:executeString(cmd);
		Logger.notice("[MindBridge] : Announce Your Name");
                cmd = "conference "..tostring(conference_name).." play "..joined_prompt;
		Logger.notice("[MindBridge] : cmd :" ..tostring(cmd).."\n");
                response = api:executeString(cmd);
		session:execute("playback", tostring(joined_prompt));
		--session:execute("playback","silence_stream://1000");
                cmd = "conference "..tostring(conference_name).." play " .. temp_dir:gsub("\\", "/") .. "/conference-"..uuid..".mp3";
                response = api:executeString(cmd);
		session:execute("playback","/tmp/conference-"..uuid..".mp3");
		--session:execute("playback","silence_stream://500")
		--if(member_count ~= 0) then
		--	session:execute("playback", "/usr/local/freeswitch/prompt/beep.wav");
		--end
		member_type = session:getVariable("member_type");
                conference_started = session:getVariable("conference_started");
                Logger.notice("[MindBridge] : Member_type: "..tostring(member_type).." conference_started: "..tostring(conference_started).."\n");
                if (tostring(conference_started) == "true") then
                        --if (tostring(member_type) ~= "moderator") then
                                --session:execute("playback", "/usr/local/freeswitch/prompt/beep.wav");
				cmd = "conference "..tostring(conference_name).." play /usr/local/freeswitch/prompt/beep.wav"
				response = api:executeString(cmd);
                        --end
                end]]--
		cmd="curl --location --request GET 'http://localhost:10000/queue/entry/"..tostring(uuid).."/"..tostring(conference_name).."'"
                freeswitch.consoleLog("notice","[MindBridge] : API : "..tostring(cmd).."");
                curl_response = api:execute("system", cmd);
                freeswitch.consoleLog("notice","[MindBridge] :curl_response : "..tostring(curl_response).."");


        end

        function join_conference_call(conf_data, conference_name, announce_prompt_name)
                if(announce_prompt_name ~= nil) then
                        session:streamFile(announce_prompt_name);
                end

                local conference_server_ip = freeswitch.getGlobalVariable("conference_server_ip")
                conference_status = session:getVariable("conference_status");
                if(conference_status == "false") then
                        create_conference_api(conference_server_ip, conference_name)
                end

                conference_record = session:getVariable("conference_record");
                wait_mod = session:getVariable("wait_mod");
                member_type = session:getVariable("member_type");
                aleg_uuid = session:getVariable("uuid");
                --record the conference
               if (conference_record == "true") then
                        if (wait_mod == "true" and member_type ~= "moderator") then
                                Logger.notice("[MindBridge] : Call will not be recording");
                                --don't start recording yet
                        else
                                Logger.notice("[MindBridge] : Call will be recording");
                                cmd="sched_api +4 none lua "..script_dir.."/start_recording.lua "..tostring(conference_name).." "..tostring(aleg_uuid)
                                api:executeString(cmd);
                        end
                end

                active_conference_server_ip = session:getVariable("active_conference_server_ip");
                if((active_conference_server_ip == conference_server_ip) or (active_conference_server_ip == nil)) then
                       	member_count = members_counts(conference_name)
			Logger.notice("[MindBridge] : member_count :: "..tostring(member_count));

			conference_started = session:getVariable("conference_started");
                        if(tonumber(member_count) == 0 and member_type ~= "moderator") then
				session:execute("playback", sounds_dir.."/45_1.wav");
			end
			max_members = session:getVariable("max_members");
			if((tonumber(member_count) >=  tonumber(max_members)) and (tonumber(max_members) ~= 0) ) then
				session:sleep(500);
				session:streamFile(sounds_dir.."/52.mp3");
				session:sleep(500);
				try = MaxTry
				session:hangup();
				return 0;
			end
			
			if(member_type == "moderator" and tonumber(member_count) >= 0) then
				Logger.notice("[MindBridge] : Call will be recording");
                  	      	cmd="sched_api +5 none lua "..script_dir.."/moh.lua "..tostring(conference_name).." "..tostring(aleg_uuid)
                        	api:executeString(cmd);
				-- session:execute("playback", "/usr/local/freeswitch/prompt/MOH.wav");

                                --[[Logger.notice("[MindBridge] : conference_name " .. conference_name .. "");
				MOH = require 'moh';
				MOH(session, conference_name, aleg_uuid);]]--
			end

			if(member_type == "moderator") then
				conference_name = session:getVariable("conference_name");
                                Logger.notice("[MindBridge] : Moderator Joining conference_name : " ..conference_name .. "\n");
                                cmd="conference "..conference_name.." set caller_id_name MindBridge"
                                api:executeString(cmd);
                        end
--			conference 21234004321_mindbridge get caller_id_name
                        --session:execute("conference", conf_data);

                        --session:execute("conference", conf_data);
			conference_continuation	= session:getVariable("conference_continuation");
			--if(conference_continuation == "notallow") then
			if(conference_continuation == "f") then
				Logger.notice("[MindBridge] : Conference Continuation FALSE");
                                cmd="conference "..tostring(conference_name).." kill all"
				Logger.notice("[MindBridge] : API : "..tostring(cmd));
                                api:executeString(cmd);
			end
                        session:execute("conference", conf_data);

                        --member_count = members_counts(conference_name)
                        --if(tonumber(member_count) == 0) then
                        --        delete_conference_api(conference_server_ip, conference_name)
                        --end
                else
                        destination_number = session:getVariable("did_number");
                        Logger.notice("[MindBridge] : Conference ID ["..tostring(conference_name).."] Active On Server ["..tostring(active_conference_server_ip).." destination_number : " ..tostring(destination_number).."]");
                        --destination_number = session:getVariable("destination_number")
			--Logger.notice("[MindBridge] : destination_number : " ..tostring(destination_number).."");

                        session:execute("bridge","{sip_h_X-conference_name="..tostring(conference_name)..",sip_h_X-conference-ID="..tostring(conf_data).."}sofia/external/"..tostring(destination_number).."@"..tostring(active_conference_server_ip));
                        session:hangup();
                        return 0;
                end
        end

        function play_member_count(announce_count, member_count, member_count_prompt)
                --play member count
                if (announce_count == "true") then
                	--if (member_count == "2") then
                        --there is one other member in this conference
                          --      session:execute("playback", "/usr/local/freeswitch/prompt/there_is_one.wav");
                        if (member_count == "1") then
--                              if (wait_mod == "true" and member_type ~= "moderator") then
                                session:execute("playback", "/usr/local/freeswitch/prompt/there_is_one.wav");
                                      --  session:execute("playback", member_count_prompt);
--                              end
                                --conference profile defines the alone sound file
                        else
                                session:execute("playback", "/usr/local/freeswitch/prompt/there-are.wav");
                                --say the count
                                session:execute("say", "en number pronounced "..member_count);
                                --members in this conference
                                session:execute("playback", "/usr/local/freeswitch/prompt/party_conf.wav");
                        end

		end
        end

        --=================routing started========================


        if (session:ready()) then
                session:preAnswer();
                session:sleep(1000);

                sounds_dir = session:getVariable("prompt_dir");
                Enable = session:getVariable("Enable");
                Logger.info("[MindBridge] Enable: " .. tostring(Enable) .. "");
                domain_name = session:getVariable("domain_name");
                --destination_number = session:getVariable("sip_to_user");
                destination_number = session:getVariable("destination_number");
                session:setVariable("did_number", tostring(destination_number));
		caller_id_number = session:getVariable("caller_id_number");
		session:setVariable("from_number", tostring(caller_id_number));

                --caller_id_number = session:getVariable("caller_id_number");
                conference_ID = session:getVariable("sip_h_X-conference-ID");

--[[                if(conference_ID ~= nil) then
                        conference_name = session:getVariable("sip_h_X-conference_name");
                        session:execute("conference", conference_ID);
                        member_count = members_counts(conference_name)
                        if(tonumber(member_count) == 0) then
                                local conf_ip = freeswitch.getGlobalVariable("conference_server_ip")
                                delete_conference_api(conf_ip, conference_name)
                        end

                        session:hangup();
                end
]]--
                Logger.info("[MindBridge] destination_number: " .. destination_number .. "");
                Logger.info("[MindBridge] caller_id_number: " .. caller_id_number .. "");

                wait_mod = "true";
                enter_sound = "tone_stream://v=-20;%(100,1000,100);v=-20;%(90,60,440);%(90,60,620)";
                exit_sound = "tone_stream://v=-20;%(90,60,620);/%(90,60,440)";

                APIcmd = "curl --location --request GET '"..tostring(UrL).."did/"..tostring(destination_number).."'"
                Logger.info ("[MindBridge] : APIcmd : "..tostring(APIcmd).."");
                curl_response = api:execute("system", APIcmd);
                Logger.debug ("[MindBridge] : API Response : "..tostring(curl_response).."");

                local encode = json.decode (curl_response)
                local Msg = encode['Msg']
                Logger.debug ("[MindBridge] : API Msg : "..tostring(Msg).."");
                if(Msg == "Failed") then
                        local ErrorMsg = encode['Error']
                        Logger.error ("[MindBridge] : ErrorMsg : "..tostring(ErrorMsg));
                        session:hangup();
                        return 0;
                end

                local solution_type = encode['did']['solution_type']
                local pinless = encode['did']['pinless']
                local welcome_music = encode['did']['welcome_music']
                local retry_music = encode['did']['retry_music']
                local wrong_pin_music = encode['did']['wrong_pin_music']
                local after_pin_music = encode['did']['after_pin_music']
                local joining_music = encode['did']['joining_music']

                if(welcome_music == "" or welcome_music == nil or welcome_music == "nil") then
                        welcome_music = nil;
                end
                if(retry_music == "" or retry_music == nil or retry_music == "nil") then
                        retry_music = nil;
                end
                if(wrong_pin_music == "" or wrong_pin_music == nil or wrong_pin_music == "nil") then
                        wrong_pin_music = nil;
                end
                if(after_pin_music == "" or after_pin_music == nil or after_pin_music == "nil") then
                        after_pin_music = nil;
                end
                if(joining_music == "" or joining_music == nil or joining_music == "nil") then
                        joining_music = nil;
                end

                Logger.debug ("[MindBridge] : solution_type : "..tostring(solution_type));
                Logger.debug ("[MindBridge] : pinless : "..tostring(pinless));
                session:setVariable("solution_type", tostring(solution_type));


                --solution type 'Mindbridge'
                if (solution_type == "Mindbridge") then
                        Logger.info("[MindBridge] : : proceed for solution type " .. tostring(solution_type) .. "");
                        profile = "Mindbridge";
                        if(pinless == "false") then
                                if(welcome_music) then
                                        conference_greeting = welcome_music;
                                else
                                        conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                end

--                              conference_greeting = sounds_dir.."/1.mp3";
                                if (session:ready()) then
                                        Logger.debug ("[MindBridge] : Try-1 pin verification");
                                        pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                else
                                        session:hangup();
                                        return 0;
                                end

				if( pin_number == "not found") then
                                                Logger.debug ("[MindBridge] : Pin not valid");
				                conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-2 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
						--Not_Valid_Pin(conference_greeting)
				elseif (pin_number == nil) then

                                       if(retry_music) then
                                              conference_greeting = retry_music
                                        else
				                conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
			   	
                                       end
			

                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-2 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end
----------------------------------------------------------------------------------------------------------------------------------------------------		 
                               if (pin_number == nil) then
                                        if(retry_music) then
                                                conference_greeting = retry_music
                                        else
						 conference_greeting = sounds_dir.."/invalid_pin.wav"

						  conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						-- conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-3 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end
				 if (pin_number == nil) then
                                        if(retry_music) then
                                                conference_greeting = retry_music

                                        else
						 conference_greeting = sounds_dir.."/invalid_pin.wav"
                   
		                    	--	 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						-- conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

--                                      conference_greeting = sounds_dir.."/31.mp3"
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-4 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end

				 if (pin_number == nil) then
                                        if(retry_music) then
                                                conference_greeting = retry_music
                                        else
					--	 conference_greeting = sounds_dir.."/invalid_pin.wav"

			                 	 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						-- conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

--                                  conference_greeting = sounds_dir.."/31.mp3"
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-5 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end

                        else
                                Logger.info("[MindBridge] conference is pinless");
                                pin_number = "verified";
                        end

                        if(pin_number == "verified") then
                                Logger.info("[MindBridge] PIN verified successfully");
				enter_sound = session:getVariable("enter_sound");
				exit_sound = session:getVariable("exit_sound");
                                aleg_uuid = session:getVariable("uuid");
                                conference_name = conference_id
				session:setVariable("conference_name", tostring(conference_name))
                                memberID, is_moderator, isLock = ModeratoriSOn(aleg_uuid, conference_name);
                                announce_name = session:getVariable("announce_name");
                                member_count = members_counts(conference_name)
                                Logger.info("[MindBridge] : conference_name : "..tostring(conference_name).." is_moderator : "..tostring(is_moderator).."\n");
				if (tonumber(memberID) == 0 and member_type == "moderator" and tonumber(member_count) == 0) then
                                        session:streamFile(sounds_dir.."/c234.wav");
				end
				enable = session:getVariable("enable_disable");
                                if (announce_name == "true") then
                                        record_name_prompt = sounds_dir.."/031.wav";
                                        record_your_name(record_name_prompt)

					os.execute("lame --scale 5 /tmp/conference-/"..aleg_uuid..".mp3 /tmp/"..aleg_uuid..".mp3")
					Logger.notice("[MindBridge] : Result after lame \n");
					if (tostring(enable) == "f") then

                			conference_started = session:getVariable("conference_started");
                                	Logger.info("[MindBridge] : id_moderator : "..tostring(is_moderator).."");

					if (member_type == "moderator" and tonumber(member_count) > 0) then
                                        if (announce_name == "true") then
                                                joined_prompt = sounds_dir.."/04.mp3";
                                                announce_your_name(conference_name, joined_prompt)
                                        else
                                                if (isLock == "false" or isLock == false) then
                                                        if (sounds == "true") then
                                                                cmd = "conference "..tostring(conference_name).." play "..enter_sound;
                                                                response = api:executeString(cmd);
                                                        end
                                                end
					end
					end
					if (is_moderator == "true" and member_type == "participant") then
                                        if (announce_name == "true") then
                                                joined_prompt = sounds_dir.."/04.mp3";
                                                announce_your_name(conference_name, joined_prompt)
                                        else
                                                if (isLock == "false" or isLock == false) then
                                                        if (sounds == "true") then
                                                                cmd = "conference "..tostring(conference_name).." play "..enter_sound;
                                                                response = api:executeString(cmd);
                                                        end
                                                end
					end
                                        end
					end
                                end
                                --get the conference member count
                                --member_count = members_counts(conference_name)
                                max_members = session:getVariable("max_members");
                                if((tonumber(member_count) >=  tonumber(max_members)) and (tonumber(max_members) ~= 0) ) then
                                        session:sleep(500);
                                        session:streamFile(sounds_dir.."/52.mp3");
                                        session:sleep(500);
                                        try = MaxTry
                                        session:hangup();
                                        return 0;
                                end

                                Logger.notice("[MindBridge] :  member_count : ".. tostring(member_count) .."");
                                wait_mod = session:getVariable("wait_mod");
				if (member_type ~= "moderator" and tonumber(member_count) ~= 0) then
                                	play_member_count(announce_count, member_count, sounds_dir.."/45_1.wav")
				end
				if (member_type == "moderator") then
                                        flags = flags .. "moderator|endconf";
--                                         if (moderator_endconf == "true") then
--                                                 flags = flags .. "|endconf";
--                                         end
                                end

                                if(session:ready()) then
			if(member_type == "moderator" and tonumber(member_count) >= 0) then
			--[[	Logger.notice("[MindBridge] : MOH of for moderator");
                  	      cmd="sched_api +5 none lua "..script_dir.."/moh.lua "..tostring(conference_name).." "..tostring(aleg_uuid)
                        	api:executeString(cmd);
				-- session:execute("playback", "/usr/local/freeswitch/prompt/MOH.wav");


                                Logger.notice("[MindBridge] : conference_name " .. conference_name .. "");
				MOH = require 'moh';
				MOH(session, conference_name, aleg_uuid);]]--
			end
                                        cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
                                        Logger.info("[MindBridge] : conference " .. cmd .. "");
                                        join_conference_call(cmd, conference_name, nil)
                                end
                                        session:hangup();
                                        return 0;
                        else
				conference_greeting = sounds_dir.."/invalid_pin.wav"
                                session:streamFile(sounds_dir.."/77.wav");
				session:streamFile(sounds_dir.."/option_not_available.wav");
                                session:hangup();
                                return 0;
                        end

                elseif (solution_type == "All") then
                        profile = "default";
                        Logger.debug ("[MindBridge] : MindBridge Conf Service Type : All");

                        if(pinless == "false") then
                                if(welcome_music) then
                                        conference_greeting = welcome_music;
                                else
                                     --   conference_greeting = sounds_dir.."/2.mp3";
					 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                end
--                              conference_greeting = sounds_dir.."/2.mp3";
                                if (session:ready()) then
                                        Logger.debug ("[MindBridge] : Try-1 pin verification");
                                        pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                else
                                        session:hangup();
                                        return 0;
                                end
                                if (pin_number == nil) then
					conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                        if(retry_music) then
                                                conference_greeting = retry_music
                                        else
						 conference_greeting = sounds_dir.."/invalid_pin.wav"

					--	conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						--conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

--                                      conference_greeting = sounds_dir.."/31.mp3"
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-2 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end
                                if (pin_number == nil) then
					conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                        if(retry_music) then
                                                conference_greeting = retry_music
                                        else
						 conference_greeting = sounds_dir.."/invalid_pin.wav"

					--	 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						 --conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

--                                      conference_greeting = sounds_dir.."/31.mp3"
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-3 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end
				-----------------------------------------------------------------------
				 if (pin_number == nil) then
					 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                        if(retry_music) then
                                                conference_greeting = retry_music
                                        else
						 conference_greeting = sounds_dir.."/invalid_pin.wav"
					--	 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						 --conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

--                                      conference_greeting = sounds_dir.."/31.mp3"
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-4 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end

				 if (pin_number == nil) then
					conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
                                        if(retry_music) then
                                                conference_greeting = retry_music
                                        else
						 conference_greeting = sounds_dir.."/invalid_pin.wav"

					--	 conference_greeting = sounds_dir.."/1Mindbridge-Welcome.wav";
						-- conference_greeting = sounds_dir.."/invalid_pin.wav"
                                        end

--                                      conference_greeting = sounds_dir.."/31.mp3"
                                        if (session:ready()) then
                                                Logger.debug ("[MindBridge] : Try-5 pin verification");
                                                pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
                                        else
                                                session:hangup();
                                                return 0;
                                        end
                                end

                        else
                                Logger.info("[MindBridge] conference is pinless");
                                pin_number = "verified";
                        end

                        if (pin_number == "verified") then
				enter_sound = session:getVariable("enter_sound");
				exit_sound = session:getVariable("exit_sound");
                                aleg_uuid = session:getVariable("uuid");
                                conference_name = conference_id
                                Logger.info("[MindBridge] :conference_name : "..tostring(conference_name).."");
                                memberID, is_moderator, isLock = ModeratoriSOn(aleg_uuid, conference_name);

                                announce_name = session:getVariable("announce_name");
                                if (announce_name == "true") then
                                        record_name_prompt = sounds_dir.."/24.mp3";
                                        record_your_name(record_name_prompt)

                                        if (announce_name == "true") then
                                                joined_prompt = sounds_dir.."/63.mp3";
                                                announce_your_name(conference_name, joined_prompt)
                                        else
--                                              if (not conference_locked) then
                                                        if (sounds == "true") then
                                                                cmd = "conference "..tostring(conference_name).." play "..enter_sound;
                                                                response = api:executeString(cmd);
                                                        end
--                                              end
                                        end
                                end

                                --get the conference member count
                                member_count = members_counts(conference_name)
                                max_members = session:getVariable("max_members");
                                if((tonumber(member_count) >=  tonumber(max_members)) and (tonumber(max_members) ~= 0) ) then
                                        session:sleep(500);
                                        session:streamFile(sounds_dir.."/52.mp3");
                                        session:sleep(500);
                                        try = MaxTry
                                        session:hangup();
                                        return 0;
                                end

                                Logger.notice("[MindBridge] :  member_count : ".. tostring(member_count) .."");
                                wait_mod = session:getVariable("wait_mod");
                                --play member count
                                play_member_count(announce_count, member_count, sounds_dir.."/45_1.wav")

                                --TODO need to verify flow
                                Logger.info("[MindBridge] : conference_name : "..tostring(conference_name).."");
                                if(session:ready()) then
			if(member_type == "moderator" and tonumber(member_count) >= 0) then
				Logger.notice("[MindBridge] : Call will be recording");
                  	      cmd="sched_api +5 none lua "..script_dir.."/moh.lua "..tostring(conference_name).." "..tostring(aleg_uuid)
                        	api:executeString(cmd);
				-- session:execute("playback", "/usr/local/freeswitch/prompt/MOH.wav");

			end
                                        cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
                                        Logger.info("[MindBridge] : conference " .. cmd .. "");
                                        join_conference_call(cmd, conference_name, sounds_dir.."/55.mp3")
                                end
                                session:hangup();
                                return 0;

                        else
                                session:streamFile(sounds_dir.."/77.wav");
				session:streamFile(sounds_dir.."/option_not_available.wav");
                                session:hangup();
                                return 0;
                        end

                else
                        Logger.info("[MindBridge] : : proceed for solution type " .. tostring(solution_type) .. "");
                        session:hangup();
                end
      end

