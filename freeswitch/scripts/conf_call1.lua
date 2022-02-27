--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  conf_call.lua --> conference controls
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

--      dofile("/usr/share/freeswitch/scripts/logger.lua");
        api = freeswitch.API();

--      package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
--      package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        xml = require "xml";
        freeswitch.consoleLog("notice", "[ConfAPI] :  Allow/disallow conference continuation");
        did_number = session:getVariable("did_number");
        conf_id = session:getVariable("conf_id");
--	sounds_dir =freeswitch.getGlobalVariable("prompt_dir")
--	session:execute("playback", sounds_dir.."/101.wav");
	--conference_name = argv[1]
	--CallUUID = argv[2]
	UrL = freeswitch.getGlobalVariable("UrL")
       	conference_name = session:getVariable("conference_name") 
	--cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

        conference_continuation = session:getVariable("conference_continuation") or 'allow';

	payload = '{"did_number":"'..tostring(did_number)..'","conference_id":"'..tostring(conf_id)..'","announce_name":"","exit_announce":"","entry_announce":"","chairperson_pin":"","conference_continuation":"'..tostring(conference_continuation)..'"}'
	APIcmd = "curl --location --request PUT '"..tostring(UrL).."conference-properties' -H 'content-type: application/json'  -d '"..tostring(payload).."'"
	Logger.info ("[MindBridge] : update conference-properties Request : "..tostring(APIcmd).."");
	curl_response = api:execute("system", APIcmd);
	Logger.debug ("[MindBridge] : update conference-properties Response : "..tostring(curl_response).."");


	if(conference_continuation == "allow") then
		endconference_grace_time = "2"
		sounds_dir =freeswitch.getGlobalVariable("prompt_dir")
                session:execute("playback", sounds_dir.."/101.wav");
		session:setVariable("conference_continuation", "notallow");

	else
		sounds_dir =freeswitch.getGlobalVariable("prompt_dir")
        session:execute("playback", sounds_dir.."/102.wav");
		endconference_grace_time = "36000"
		session:setVariable("conference_continuation", "allow");

	end
	
	cmd = "conference " .. conference_name .. " set  endconference_grace_time "..endconference_grace_time
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));


  --      api:executeString("msleep 1000")
	--reply = api:executeString("luarun keypad_commands.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
		cnt = 0	
		session:sleep(1000);
		Keypad = require 'keypad_commands-loop'
		Keypad(session, conference_name, aleg_uuid, cnt)
