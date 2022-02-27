--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  lecture_mode.lua --> conference controls
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
--      package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
--      package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        xml = require "xml";

        Logger.notice ("[ConfAPI] : conference Lecture mode enabled")
        conference_name = session:getVariable("conference_name");
        CallUUID = session:getVariable("uuid");

        UrL = freeswitch.getGlobalVariable("UrL")
        sounds_dir = session:getVariable("prompt_dir");
        conference_name = session:getVariable("conference_name");
        did_number = session:getVariable("did_number");
        conf_id = session:getVariable("conf_id");

	cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

        lecture_mode = argv[3] --session:getVariable("lecture_mode");
        if (lecture_mode == "OFF" ) then
                session:setVariable("lecture_mode","false")
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " unmute non_moderator" ;
                response = api:executeString(cmd);
                freeswitch.consoleLog("notice", "[ConfAPI] :  Lecture mode Disabled API : "..tostring(cmd));
                conference_mute = "false"
        else
                session:setVariable("lecture_mode","true")
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " mute non_moderator";
                response = api:executeString(cmd);
                freeswitch.consoleLog("notice", "[ConfAPI] :  Lecture mode Enabled API : "..tostring(cmd));
                conference_mute = "true"
        end

        payload = '{"did_number":"'..tostring(did_number)..'","conference_id":"'..tostring(conf_id)..'","conference_mute":"'..tostring(conference_mute)..'",,"announce_name":"","exit_announce":"","entry_announce":"","chairperson_pin":""}'
        APIcmd = "curl --insecure --location --request PUT '"..tostring(UrL).."conference-properties' -H 'content-type: application/json'  -d '"..tostring(payload).."'"
        Logger.info ("[MindBridge] : update conference-properties Request : "..tostring(APIcmd).."");
        curl_response = api:execute("system", APIcmd);
        Logger.debug ("[MindBridge] : update conference-properties Response : "..tostring(curl_response).."");


        -- while unmute need to check chairperson lecture_mode and base on that need to allow unmute.
