--[[
  The Initial Developer of the Original Code is
  
  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  app.lua --> conference centers 
--]]


	dofile("/usr/share/freeswitch/scripts/logger.lua");

--general functions
	require "resources.functions.base64";
	require "resources.functions.trim";
	require "resources.functions.file_exists";
	require "resources.functions.explode";
	require "resources.functions.format_seconds";
	require "resources.functions.mkdir";
	local json = require "resources.functions.lunajson"
	
	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
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
			
			Logger.info("[MindBridge] Playing file name : " .. tostring(prompt_audio_file) .. "");
			pin_number = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", prompt_audio_file, "", "\\d+");
			--remove non numerics
			pin_number = pin_number:gsub("%p","")
			
			solution_type = session:getVariable("solution_type");
			if (pin_number ~= nil) then
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
 
-- 					session:execute("say","en name_spelled pronounced " .. pin_number)
-- 					Logger.error ("[MindBridge] : API MessageRes : "..tostring(MessageRes).." "..tostring(Msg).."");
					pin_number = nil;
				else
					conf_id = encode['did_conference_map']['conference_id']
					session:setVariable("conf_id", tostring(conf_id));
					pin_number = "verified";
					chairperson_pin = encode['did_conference_map']['chairperson_pin']
					active_conference_server_ip = encode['did_conference_map']['conference_server_ip']
					member_type = encode['did_conference_map']['pin_type']
					announce_name = encode['did_conference_map']['announce_name']
					max_members = encode['did_conference_map']['max_members']

					conference_name = encode['did_conference_map']['conference_name']
					conference_record  = encode['did_conference_map']['conference_recording']
					conference_wait_for_moderator  = encode['did_conference_map']['conference_wait_for_moderator']
					conference_mute  = encode['did_conference_map']['conference_mute ']
					
					start_time  = encode['did_conference_map']['start_time ']
					end_time  = encode['did_conference_map']['end_time ']
					Logger.debug ("[MindBridge] : start_time : " ..tostring(start_time));
					Logger.debug ("[MindBridge] : end_time : " ..tostring(end_time));

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
					end
					if(flags == nil or flags == "" or flags == "nil") then
						flags = "";
					end
					Logger.debug ("[MindBridge] : flags : " ..tostring(flags));

					session:setVariable("conference_record", tostring(conference_record));
					session:setVariable("member_type", tostring(member_type));
					session:setVariable("announce_name", tostring(announce_name));
					session:setVariable("active_conference_server_ip", tostring(active_conference_server_ip));
					if(conference_name == "" or conference_name == nil or conference_name == "nil") then
						create_uuid = api:executeString("create_uuid");
						conference_id = tostring(create_uuid).."_"..tostring(conf_id)
						session:setVariable("conference_status", "false");
					else
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
					if(after_pin_music) then
						session:streamFile(after_pin_music);
					else
						session:streamFile(sounds_dir.."/thank_you.mp3");
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
		id, hear_value, speak_value = nil, nil, nil;
		is_moderator = false;
		isLock = false;
		result = api:executeString("conference "..conf_name .. " xml_list")
		if string.match(result, "not found") then
			Logger.notice("[MindBridge] : No Active conference "..tostring(conf_name))
			result = nil;
			return 0,false, false;
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
									if( string.find(tostring(j),"<id>") ~= nil )then
										id = tostring(j):match("id>(.-)</id>")
-- 										Logger.notice("[MindBridge] : CONFID : "..tostring(id));
									elseif(string.find(tostring(j),"<is_moderator>") ~= nil ) then
										is_moderator = tostring(j):match("is_moderator>(.-)</is_moderator>")
-- 										Logger.notice("[MindBridge] : IS_MODERATOR : "..tostring(is_moderator));
									end
								end
							end
						end
					end 
				end
			end
			i = i+1;
		end
		
		Logger.notice("[MindBridge] : ID : "..tostring(id).." IS_MODERATOR : ["..tostring(is_moderator).."]");
		return id, is_moderator, isLock
	end
	
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
		session:execute("playback", "tone_stream://v=-7;%%(500,0,500.0)");
		--record the response
		max_len_seconds = 5;
		silence_threshold = "500";
		silence_secs = "3";
		session:recordFile(temp_dir:gsub("\\","/") .. "/conference-"..uuid..".mp3", max_len_seconds, silence_threshold, silence_secs);
	end

	function announce_your_name(conference_name, joined_prompt)
		cmd = "conference "..tostring(conference_name).." play " .. temp_dir:gsub("\\", "/") .. "/conference-"..uuid..".mp3";
		response = api:executeString(cmd);
		cmd = "conference "..tostring(conference_name).." play "..joined_prompt;
		response = api:executeString(cmd);
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
				cmd="sched_api +3 none lua /usr/share/freeswitch/scripts/start_recording.lua "..tostring(conference_name).." "..tostring(aleg_uuid)
				api:executeString(cmd);
			end
		end
					
		active_conference_server_ip = session:getVariable("active_conference_server_ip");
		if((active_conference_server_ip == conference_server_ip) or (active_conference_server_ip == nil)) then
			session:execute("conference", conf_data);
			member_count = members_counts(conference_name)
			if(tonumber(member_count) == 0) then
				delete_conference_api(conference_server_ip, conference_name)
			end
		else
			Logger.notice("[MindBridge] : Conference ID ["..tostring(conference_name).."] Active On Server ["..tostring(active_conference_server_ip).."]");
			destination_number = session:getVariable("destination_number");
			session:execute("bridge","{sip_h_X-conference_name="..tostring(conference_name)..",sip_h_X-conference-ID="..tostring(conf_data).."}sofia/external/"..tostring(destination_number).."@"..tostring(active_conference_server_ip));
			session:hangup();
			return 0;
		end
	end
	
	function play_member_count(announce_count, member_count, member_count_prompt)
		--play member count
		if (announce_count == "true") then
			if (member_count == "1") then
			--there is one other member in this conference
				session:execute("playback", "conference/conf-one_other_member_conference.wav");
			elseif (member_count == "0") then
-- 				if (wait_mod == "true" and member_type ~= "moderator") then
 					session:execute("playback", member_count_prompt);
-- 				end
				--conference profile defines the alone sound file
			else
				--say the count
				session:execute("say", "en number pronounced "..member_count);
				--members in this conference
				session:execute("playback", "conference/conf-members_in_conference.wav");
			end
		end
	end
	
	--=================routing started========================
	
	
	if (session:ready()) then
		session:preAnswer();
		session:sleep(1000);

		sounds_dir = session:getVariable("prompt_dir");
		domain_name = session:getVariable("domain_name");
		destination_number = session:getVariable("destination_number");
		session:setVariable("did_number", tostring(destination_number));

		caller_id_number = session:getVariable("caller_id_number");
		conference_ID = session:getVariable("sip_h_X-conference-ID");
		
		if(conference_ID ~= nil) then
			conference_name = session:getVariable("sip_h_X-conference_name");
			session:execute("conference", conference_ID);
			member_count = members_counts(conference_name)
			if(tonumber(member_count) == 0) then
				local conf_ip = freeswitch.getGlobalVariable("conference_server_ip")
				delete_conference_api(conf_ip, conference_name)
			end
			
			session:hangup();
		end
		
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
		
		if(solution_type == "Reservation Less") then
			profile = "reservation-less";
			if(pinless == "false") then
				if(welcome_music) then
					conference_greeting = welcome_music;
				else
					conference_greeting = sounds_dir.."/2.mp3";
				end
				if (session:ready()) then
					Logger.debug ("[MindBridge] : Try-1 pin verification");
					pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
				else
					session:hangup();
					return 0;
				end 
				if (pin_number == nil) then
					if(retry_music) then
						conference_greeting = retry_music
					else
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-2 pin verification");
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
						conference_greeting = sounds_dir.."/31.mp3"
					end

					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-3 pin verification");
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
				aleg_uuid = session:getVariable("uuid");
				conference_name = conference_id
				Logger.info("[MindBridge] :conference_name : "..tostring(conference_name).."");
				memberID, is_moderator, isLock = ModeratoriSOn(aleg_uuid, conference_name);
				if (is_moderator == "true" or is_moderator == true) then
					Logger.info("[MindBridge] :Moderator is Already Joined conference.\n");
					if(session:ready()) then
						cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
						Logger.info("[MindBridge] : conference " .. cmd .. "");
						join_conference_call(cmd, conference_name, sounds_dir.."/55.mp3")
					end
					session:hangup();
					return 0;
				end
			end
			
			if (pin_number == "verified") then
				min_digits = 1;
				max_digits = 1;
				max_tries = 1;
				digit_timeout = 5000;
			
				-- prompt_audio_file prompt should be to join press *
				conference_greet = sounds_dir.."/58.mp3"
				Logger.info("[MindBridge] : Playing file name : " .. tostring(conference_greet) .. "");
				user_input = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", conference_greet, "", "");
				Logger.info("[MindBridge] : user_input: " .. tostring(user_input) .. "");

				--set the member type
				if (user_input == "*") then
					member_type = "moderator";
					session:setVariable("member_type", tostring(member_type));
					if (member_type == "moderator") then
						flags = flags .. "|moderator|endconf";
					end
					
					try = 0;
					MaxTry = 3;
					while(try < MaxTry ) do
						min_digits = 2;
						max_digits = 20;
						max_tries = 1;
						digit_timeout = 5000;
					
						if(try == 0) then
							conference_greet =  sounds_dir.."/59.mp3"
						else
							conference_greet =  sounds_dir.."/31.mp3"
						end
							
						Logger.info("[MindBridge] Playing file name : " .. tostring(conference_greet) .. "");
						moderator_pin = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", conference_greet, "", "\\d+");
						moderator_pin = moderator_pin:gsub("%p","")
						Logger.info("[MindBridge] : user entered pin [ "..tostring(moderator_pin).." ] and chairperson_pin [ "..tostring(chairperson_pin).." ]");

						try = try + 1;
						if(moderator_pin == chairperson_pin) then
							moderator_verified = "true";
						end
							
						-- verify moderator pin (api call).
						if (moderator_verified == "true") then
							try = MaxTry
							Logger.info("[MindBridge] : Moderator PIN Verified.");
							-- predd *2 to change record menu and pin etc
							-- press 1 to start if conference not active/ 
							-- ask to announce name follow by #
							-- to announce mute press *6 / unmute press *6 
							-- Play Beep after announcement.

							Retry = 0;
							MaxReTry = 3;
							while(Retry < MaxReTry ) do
								min_digits_tmp = 1;
								max_digits_tmp = 2;
								max_tries_tmp = 1;
								digit_timeout_tmp = 2000;

								conference_greet =  sounds_dir.."/54.mp3"
								Logger.info("[MindBridge] Playing file name : " .. tostring(conference_greet) .. "");
								option = session:playAndGetDigits(min_digits_tmp, max_digits_tmp, max_tries_tmp, digit_timeout_tmp, "#", conference_greet, "", "\\d+");

								Retry = Retry + 1;
								if (option == "*2") then
									session:streamFile(sounds_dir.."/conf_option_change.mp3");
									session:sleep(500);
									Logger.info("[MindBridge] : user entered change menu option.");
									session:execute("ivr", "record_menu");
									-- IVR need 
									session:sleep(500);
								elseif (option == "*0") then
									Retry = MaxReTry
									session:streamFile(sounds_dir.."/thank_you.mp3");
									session:sleep(500);

									Logger.info("[MindBridge] : user entered menu option.");
								elseif (option == "1") then
									Retry = MaxReTry
									session:streamFile(sounds_dir.."/thank_you.mp3");
									session:sleep(500);

									Logger.info("[MindBridge] : user entered create conference.");
								else
									Logger.info("[MindBridge] : INVALID INPUT\n");
								end
							end
							
							announce_name = session:getVariable("announce_name");
							if (announce_name == "true") then
								record_name_prompt = sounds_dir.."/72.mp3";
								record_your_name(record_name_prompt)
							end

							--get the conference xml_list
							cmd = "conference "..tostring(conference_name).." xml_list";
							Logger.info("[MindBridge] :" .. cmd .. "");
							result = api:executeString(cmd);

							--get the content to the <conference> tag
							result = string.match(result,[[<conference (.-)>]],1);

							--get the uuid out of the xml tag contents
							if (result ~= nil) then
								conference_session_uuid = string.match(result,[[uuid="(.-)"]],1);
							end

							--log entry
							if (conference_session_uuid ~= nil) then
								Logger.info("[MindBridge] :conference_session_uuid: " .. conference_session_uuid .. "");
							end

							--set the start epoch
							start_epoch = os.time();
							announce_name = session:getVariable("announce_name");
							if (announce_name == "true") then
								joined_prompt = sounds_dir.."/63.mp3";
								announce_your_name(conference_name, joined_prompt)
							else
								if (sounds == "true") then
									cmd = "conference "..tostring(conference_name).." play "..enter_sound;
									response = api:executeString(cmd);
								end
							end
							session:sleep(500);
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

							wait_mod = session:getVariable("wait_mod");
							play_member_count(announce_count, member_count, sounds_dir.."/45.mp3")
							
							session:streamFile(sounds_dir.."/55.mp3");
							cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
							Logger.info("[MindBridge] : conference " .. cmd .. "");
							join_conference_call(cmd, conference_name, sounds_dir.."/55.mp3")
							session:hangup();
						else
							if( try >= MaxTry ) then
								session:streamFile(sounds_dir.."/73.mp3");
							end
							if(session:ready()) then
								Logger.info("[MindBridge] : Invalid moderator PIN.. Try again!!!");
								moderator_verified = "false";
							else 
								try = MaxTry
								session:hangup();
								return 0;
							end
						end
					end
					if(moderator_verified == "false") then
						session:hangup();
						return 0;
					end
				else
					member_type = "participant";
					session:setVariable("member_type", tostring(member_type));
				end
				
				if(member_type == "participant") then
					if(session:ready()) then
						aleg_uuid = session:getVariable("uuid");
						Logger.info("[MindBridge] :conference_name : "..tostring(conference_name).."");
						memberID, is_moderator, isLock = ModeratoriSOn(aleg_uuid, conference_name);
						if (is_moderator == "true" or is_moderator == true) then
							Logger.info("[MindBridge] : Leader already joined conference."); 
						else
							session:streamFile(sounds_dir.."/82.mp3");
							session:sleep(500);
						end
						
						cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
						Logger.info("[MindBridge] : conference " .. cmd .. "");
						join_conference_call(cmd, conference_name, nil)
					end
				end
				session:hangup();
				return 0;
			else
				Logger.debug ("[MindBridge] : PIN verification failed.");
				session:streamFile(sounds_dir.."/77.mp3");
				session:hangup();
				return 0;
			end
			
		--solution type 'Automated' 
		elseif (solution_type == "Automated") then  
			Logger.info("[MindBridge] : : proceed for solution type " .. tostring(solution_type) .. "");
			profile = "Automated";
			if(pinless == "false") then
				if(welcome_music) then
					conference_greeting = welcome_music;
				else
					conference_greeting = sounds_dir.."/1.mp3";
				end
				
-- 				conference_greeting = sounds_dir.."/1.mp3";
				if (session:ready()) then
					Logger.debug ("[MindBridge] : Try-1 pin verification");
					pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
				else
					session:hangup();
					return 0;
				end 
				if (pin_number == nil) then
					if(retry_music) then
						conference_greeting = retry_music
					else
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-2 pin verification");
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
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-3 pin verification");
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
				aleg_uuid = session:getVariable("uuid");
				conference_name = conference_id
				Logger.info("[MindBridge] : conference_name : "..tostring(conference_name).."");
				memberID, is_moderator, isLock = ModeratoriSOn(aleg_uuid, conference_name);
				announce_name = session:getVariable("announce_name");
				if (announce_name == "true") then
					record_name_prompt = sounds_dir.."/24.mp3";
					record_your_name(record_name_prompt)
					
					if (announce_name == "true") then
						joined_prompt = sounds_dir.."/63.mp3";
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
				play_member_count(announce_count, member_count, sounds_dir.."/45.mp3")

				if (member_type == "moderator") then
					flags = flags .. "endconf|moderator";
					if (moderator_endconf == "true") then
						flags = flags .. "|endconf";
					end
				end
				
				if(session:ready()) then
					cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
					Logger.info("[MindBridge] : conference " .. cmd .. "");
					join_conference_call(cmd, conference_name, nil)
				end
					session:hangup();
					return 0;
			else
				session:streamFile(sounds_dir.."/77.mp3");
				session:hangup();
				return 0;
			end
			
		elseif (solution_type == "Event Service") then  
			profile = "event-service";
			Logger.debug ("[MindBridge] : MindBridge Conf Service Type : Event Service");
			if(pinless == "false") then
				if(welcome_music) then
					conference_greeting = welcome_music;
				else
					conference_greeting = sounds_dir.."/2.mp3";
				end
							
-- 				conference_greeting = sounds_dir.."/2.mp3";
				if (session:ready()) then
					Logger.debug ("[MindBridge] : Try-1 pin verification");
					pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
				else
					session:hangup();
					return 0;
				end 
				if (pin_number == nil) then
					if(retry_music) then
						conference_greeting = retry_music
					else
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-2 pin verification");
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
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-3 pin verification");
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
-- 						if (not conference_locked) then
							if (sounds == "true") then
								cmd = "conference "..tostring(conference_name).." play "..enter_sound;
								response = api:executeString(cmd);
							end
-- 						end
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
				play_member_count(announce_count, member_count, sounds_dir.."/45.mp3")

				Logger.info("[MindBridge] :conference_name : "..tostring(conference_name).."");
				if(session:ready()) then
					cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
					Logger.info("[MindBridge] : conference " .. cmd .. "");
					join_conference_call(cmd, conference_name, nil)
				end
				session:hangup();
				return 0;
				
			else 
				session:streamFile(sounds_dir.."/77.mp3");
				session:hangup();
				return 0;
			end
			
		elseif (solution_type == "Event Plus") then  
			profile = "event-plus";
			Logger.debug ("[MindBridge] : MindBridge Conf Service Type : Event Plus");

			if(pinless == "false") then
				if(welcome_music) then
					conference_greeting = welcome_music;
				else
					conference_greeting = sounds_dir.."/4.mp3";
				end
							
-- 				conference_greeting = sounds_dir.."/4.mp3";
				if (session:ready()) then
					Logger.debug ("[MindBridge] : Try-1 pin verification");
					pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
				else
					session:hangup();
					return 0;
				end
				
				if (pin_number == nil) then
					Logger.debug ("[MindBridge] : No security code has detected");
					conference_greeting = sounds_dir.."/81.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Direct Dialing to Operator.");
						--TODO 
-- 						pin_number, chairperson_pin, conference_id,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
						conference_name = uuid.."-operator"
-- 						session:execute("conference","bridge:"..tostring(conference_name).."@"..tostring(profile)..":sofia/internal/9974743321@18.216.103.124:9000|endpoint[+flags{}]");
						session2 = freeswitch.Session("sofia/internal/9974743321@18.216.103.124:9000");
-- 						session:execute("bridge","sofia/internal/9974743321@18.216.103.124:9000");
						local obCause = session2:hangupCause()
						freeswitch.consoleLog("info", "obSession:hangupCause() = " .. obCause )
	
						if(obCause == "SUCCESS") then
							freeswitch.bridge(session, session2);
						elseif ( obCause == "USER_BUSY" ) then
							Logger.debug ("[MindBridge] : USER_BUSY");
							-- SIP 486
						-- For BUSY you may reschedule the call for later
						elseif ( obCause == "NO_ANSWER" ) then
							Logger.debug ("[MindBridge] : NO_ANSWER");
						-- Call them back in an hour

						else
						-- Log these issues
						end
	
						session:hangup();
						return 0;
						
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
-- 						if (not conference_locked) then
							if (sounds == "true") then
								cmd = "conference "..tostring(conference_name).." play "..enter_sound;
								response = api:executeString(cmd);
							end
-- 						end
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
				play_member_count(announce_count, member_count, sounds_dir.."/45.mp3")

				Logger.info("[MindBridge] :conference_name : "..tostring(conference_name).."");
				if(session:ready()) then
					cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
					Logger.info("[MindBridge] : conference " .. cmd .. "");
					join_conference_call(cmd, conference_name, nil)
				end
				session:hangup();
				return 0;
			else
				session:streamFile(sounds_dir.."/77.mp3");
				session:hangup();
				return 0;
			end
			
		elseif (solution_type == "Direct Event") then  
			profile = "direct-event";
			Logger.debug ("[MindBridge] : MindBridge Conf Service Type : Direct Event");
			
			if(pinless == "false") then
				if(welcome_music) then
					conference_greeting = welcome_music;
				else
					conference_greeting = sounds_dir.."/6.mp3";
				end
				
-- 				conference_greeting = sounds_dir.."/6.mp3";
				if (session:ready()) then
					Logger.debug ("[MindBridge] : Try-1 pin verification");
					pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
				else
					session:hangup();
					return 0;
				end 
				if (pin_number == nil) then
					if(retry_music) then
						conference_greeting = retry_music
					else
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-2 pin verification");
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
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-3 pin verification");
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
-- 						if (not conference_locked) then
							if (sounds == "true") then
								cmd = "conference "..tostring(conference_name).." play "..enter_sound;
								response = api:executeString(cmd);
							end
-- 						end
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
				play_member_count(announce_count, member_count, sounds_dir.."/45.mp3")
				
				Logger.info("[MindBridge] :conference_name : "..tostring(conference_name).."");
				if(session:ready()) then
-- 					session:streamFile(sounds_dir.."/55.mp3");
					cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
					Logger.info("[MindBridge] : conference " .. cmd .. "");
					join_conference_call(cmd, conference_name, sounds_dir.."/55.mp3")
				end
				session:hangup();
				return 0;
			else
				session:streamFile(sounds_dir.."/77.mp3");
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
					conference_greeting = sounds_dir.."/2.mp3";
				end
-- 				conference_greeting = sounds_dir.."/2.mp3";
				if (session:ready()) then
					Logger.debug ("[MindBridge] : Try-1 pin verification");
					pin_number, chairperson_pin, conference_id, member_type,flags = get_pin_number(conference_greeting, wrong_pin_music, after_pin_music);
				else
					session:hangup();
					return 0;
				end 
				if (pin_number == nil) then
					if(retry_music) then
						conference_greeting = retry_music
					else
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-2 pin verification");
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
						conference_greeting = sounds_dir.."/31.mp3"
					end
					
-- 					conference_greeting = sounds_dir.."/31.mp3"
					if (session:ready()) then
						Logger.debug ("[MindBridge] : Try-3 pin verification");
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
-- 						if (not conference_locked) then
							if (sounds == "true") then
								cmd = "conference "..tostring(conference_name).." play "..enter_sound;
								response = api:executeString(cmd);
							end
-- 						end
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
				play_member_count(announce_count, member_count, sounds_dir.."/45.mp3")
				
				--TODO need to verify flow
				Logger.info("[MindBridge] : conference_name : "..tostring(conference_name).."");
				if(session:ready()) then
					cmd = ""..tostring(conference_name).."@"..profile.."+flags{".. tostring(flags) .."}";
					Logger.info("[MindBridge] : conference " .. cmd .. "");
					join_conference_call(cmd, conference_name, sounds_dir.."/55.mp3")
				end
				session:hangup();
				return 0;
				
			else
				session:streamFile(sounds_dir.."/77.mp3");
				session:hangup();
				return 0;
			end

		else
			Logger.info("[MindBridge] : : proceed for solution type " .. tostring(solution_type) .. "");
			session:hangup();
		end
	end
	