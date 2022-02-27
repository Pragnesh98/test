--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  disconnect_all_line.lua --> conference controls
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

        freeswitch.consoleLog("notice", "[ConfAPI] :  Disconnect all lines except leader's - Leader Only");
        sounds_dir = session:getVariable("prompt_dir");
        conference_name = session:getVariable("conference_name");
        CallUUID = session:getVariable("uuid");
        cmd = "conference " .. conference_name .. " hup non_moderator"
        response = api:executeString(cmd);
        freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
	
        cmd = "conference " .. conference_name .. " moh off"
        response = api:executeString(cmd);
        freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
	 session:execute("playback", sounds_dir.."/disconnect-call.wav");
