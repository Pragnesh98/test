--set variables
	flags = "";
	max_tries = 3;
	digit_timeout = 5000;

	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
	xml = require "xml";
	
	dofile("/usr/share/freeswitch/scripts/logger.lua");

--general functions
	require "resources.functions.base64";
	require "resources.functions.trim";
	require "resources.functions.file_exists";
	require "resources.functions.explode";
	require "resources.functions.format_seconds";
	require "resources.functions.mkdir";
	local json = require "resources.functions.lunajson"
	
	debug["debug"] = false;
	
--prepare the api object
	api = freeswitch.API();

	temp_dir = "/tmp/"
	flags = ""
	UrL = "http://13.234.67.212:10000/"

	uuid = session:getVariable("uuid");
	session:answer();
	
	function get_pin_number(prompt_audio_file)
		if (session:ready()) then
			min_digits = 2;
			max_digits = 20;
			max_tries = 1;
			digit_timeout = 5000;
			
			Logger.info("[MindBridge] Playing file name : " .. tostring(prompt_audio_file) .. "");
			pin_number = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", prompt_audio_file, "", "\\d+");
			--remove non numerics
			pin_number = pin_number:gsub("%p","")
			
			if (pin_number ~= nil) then
				APIcmd = "curl --location --request GET '"..tostring(UrL).."did/verify-pin/"..tostring(destination_number).."/"..tostring(pin_number).."'"
				Logger.info ("[MindBridge] : Conf PIN verification APIcmd : "..tostring(APIcmd).."");
				curl_response = api:execute("system", APIcmd);
				Logger.debug ("[MindBridge] : Conf PIN verification API Response : "..tostring(curl_response).."");

				local encode = json.decode (curl_response)
				local MessageRes = encode['message']
				if(MessageRes == "Not Found") then
					Logger.error ("[MindBridge] : API MessageRes : "..tostring(MessageRes).."");
					pin_number = nil;
				else
					conf_id = encode['did_conference_map']['conference_id']
					session:setVariable("conf_id", tostring(conf_id));
					conference_id = tostring(destination_number).."_"..tostring(pin_number).."_"..tostring(conf_id)
					pin_number = "verified";
					chairperson_pin = encode['did_conference_map']['chairperson_pin']
					conference_server_ip = encode['did_conference_map']['chairperson_pin']
					
					Logger.debug ("[MindBridge] : conference_server_ip : " ..tostring(conference_server_ip));
					Logger.debug ("[MindBridge] : chairperson_pin : " ..tostring(chairperson_pin));
					Logger.debug ("[MindBridge] : Conf PIN verified Successfully");
					session:streamFile(sounds_dir.."/thank_you.mp3");
					session:sleep(500);
				end
			else
				pin_number = nil;
				conference_id = nil;
			end

			if (pin_number == nil) then
				return nil, nil, nil;
			else
				Logger.info("[MindBridge] Conf PIN : " .. tostring(pin_number) .. "");
				return pin_number, chairperson_pin, conference_id;
			end
		else
			session:hangup();
			return 0;
		end
	end


		sounds_dir = session:getVariable("prompt_dir");
		domain_name = session:getVariable("domain_name");
		destination_number = session:getVariable("destination_number");
		caller_id_number = session:getVariable("caller_id_number");
		
		Logger.info("[MindBridge] destination_number: " .. destination_number .. "");
		Logger.info("[MindBridge] caller_id_number: " .. caller_id_number .. "");
		
	
