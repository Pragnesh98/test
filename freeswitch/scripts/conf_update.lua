--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  conf_update.lua --> conference centers
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

        xml = require "xml";

        debug["debug"] = false;
        api = freeswitch.API();

        Logger.notice ("[MindBridge] : Conf Menu Change Request");

        request_type = argv[1]
        user_input = argv[2]

        Logger.info ("[MindBridge] : request_type : "..tostring(request_type).."");

        UrL = freeswitch.getGlobalVariable("UrL")
        sounds_dir = session:getVariable("prompt_dir");
        conference_name = session:getVariable("conference_name");
        did_number = session:getVariable("did_number");
        conf_id = session:getVariable("conf_id");

        Logger.info ("[MindBridge] : conference_name : "..tostring(conference_name).."");
        Logger.info ("[MindBridge] : did_number : "..tostring(did_number).."");
        Logger.info ("[MindBridge] : conf_id : "..tostring(conf_id).."");

        cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

        if(request_type == "PIN") then
                Logger.info ("[MindBridge] : Update Conference PIN ["..tostring(user_input).."]");
                chairperson_pin = user_input
                entry_announce = ""
                announce_name = ""
                exit_announce = ""

                payload = '{"did_number":"'..tostring(did_number)..'","conference_id":"'..tostring(conf_id)..'","announce_name":"'..tostring(announce_name)..'","exit_announce":"'..tostring(exit_announce)..'","entry_announce":"'..tostring(entry_announce)..'","chairperson_pin":"'..tostring(chairperson_pin)..'"}'
                APIcmd = "curl --insecure --location --request PUT '"..tostring(UrL).."conference-properties' -H 'content-type: application/json'  -d '"..tostring(payload).."'"
                Logger.info ("[MindBridge] : update conference-properties Request : "..tostring(APIcmd).."");
                curl_response = api:execute("system", APIcmd);
                Logger.debug ("[MindBridge] : update conference-properties Response : "..tostring(curl_response).."");

                session:streamFile(sounds_dir.."/your_leader_pin_is.mp3");
                session:execute("say","en name_spelled pronounced " .. user_input)

        elseif(request_type == "OPTION") then
                Logger.info ("[MindBridge] : Conference OPTION Change Request");
                min_digits = 1;
                max_digits = 1;
                max_tries = 1;
                digit_timeout = 5000;

                conference_greet = sounds_dir.."/announce_option_1_and_2.mp3"
                Logger.info("[MindBridge] : Playing file name : " .. tostring(conference_greet) .. "");
                option = session:playAndGetDigits(min_digits, max_digits, max_tries, digit_timeout, "#", conference_greet, "", "");
                Logger.info("[MindBridge] : user_input: " .. tostring(option) .. "");

                if(option == "1") then
                        Logger.info("[MindBridge] : selected option Announce Name");
                        enter_sound = "announce_name"
                        exit_sound = "announce_name"
                elseif(option == "2") then
                        Logger.info("[MindBridge] : selected option Announce Tone");
                        enter_sound = "tone";
                        exit_sound = "tone";
                elseif(option == "3") then
                        Logger.info("[MindBridge] : selected option Silence");
                        enter_sound = "silence";
                        exit_sound = "silence";
                else
                        Logger.info("[MindBridge] : Invalid Input");
                end
                chairperson_pin = ""
                announce_name = ""

                payload = '{"did_number":"'..tostring(did_number)..'","conference_id":"'..tostring(conf_id)..'","announce_name":"'..tostring(announce_name)..'","exit_announce":"'..tostring(exit_announce)..'","entry_announce":"'..tostring(entry_announce)..'","chairperson_pin":"'..tostring(chairperson_pin)..'"}'
                APIcmd = "curl --insecure --location --request PUT '"..tostring(UrL).."conference-properties' -H 'content-type: application/json'  -d '"..tostring(payload).."'"
                Logger.info ("[MindBridge] : update conference-properties Request : "..tostring(APIcmd).."");
                curl_response = api:execute("system", APIcmd);
                Logger.debug ("[MindBridge] : update conference-properties Response : "..tostring(curl_response).."");

        elseif(request_type == "ANNOUNCE_NAME") then
                announce_name = session:getVariable("announce_name");
                chairperson_pin = ""
                entry_announce = ""
                exit_announce = ""

                if(announce_name == "true") then
                        Logger.info ("[MindBridge] : Conference ANNOUNCE_NAME Change Request");
                        announce_name = "false"   --toggle
                else
                        announce_name = "true"   --toggle
                end

                payload = '{"did_number":"'..tostring(did_number)..'","conference_id":"'..tostring(conf_id)..'","announce_name":"'..tostring(announce_name)..'","exit_announce":"'..tostring(exit_announce)..'","entry_announce":"'..tostring(entry_announce)..'","chairperson_pin":"'..tostring(chairperson_pin)..'"}'
                APIcmd = "curl --insecure --location --request PUT '"..tostring(UrL).."conference-properties' -H 'content-type: application/json'  -d '"..tostring(payload).."'"
                Logger.info ("[MindBridge] : update conference-properties Request : "..tostring(APIcmd).."");
                curl_response = api:execute("system", APIcmd);
                Logger.debug ("[MindBridge] : update conference-properties Response : "..tostring(curl_response).."");

                if(announce_name == "true") then
                        session:sleep(500);
                        session:streamFile(sounds_dir.."/announce_name_on.mp3");
                        session:sleep(500);
                else
                        session:sleep(500);
                        session:streamFile(sounds_dir.."/announce_name_off.mp3");
                        session:sleep(500);
                end

        elseif (request_type == "AUTO") then
                Logger.info ("[MindBridge] : Conference AUTO Continues Setting Change Request");
        else
                Logger.info ("[MindBridge] : Invalid request");
                session:hangup();
                return 0;
        end
