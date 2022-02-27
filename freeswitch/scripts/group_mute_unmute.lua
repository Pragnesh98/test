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


        freeswitch.consoleLog("notice", "[ConfAPI] :  Group Mute/Unmute");
        conference_name = session:getVariable("conference_name");
        CallUUID = session:getVariable("uuid");

	cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

        group_mute_state = argv[3] --session:getVariable("group_mute_state");
        if (group_mute_state == "unmute-all" ) then
                session:setVariable("group_mute_state","false")
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " unmute non_moderator" ;
                response = api:executeString(cmd);

                freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
                freeswitch.consoleLog("notice", "[ConfAPI] :  Group unmuted");
		session:execute("playback", "/usr/local/freeswitch/prompt/unmute.wav");

        else
                session:setVariable("group_mute_state","true")
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " mute non_moderator";
                response = api:executeString(cmd);

                freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
                freeswitch.consoleLog("notice", "[ConfAPI] :  Group muted");
		session:execute("playback", "/usr/local/freeswitch/prompt/mute.wav");

        end
            solution_type = session:getVariable("solution_type");   
        if(solution_type == "Mindbridge") then
                cmd = "conference " .. conference_name .. " play /usr/local/freeswitch/prompt/Mute_all_paty.wav " ..tostring(MemberID);
                freeswitch.consoleLog("notice", "[ConfAPI] : cmd : ".. tostring(cmd) .. "")
                response = api:executeString(cmd);
	elseif(solution_type == "Mindbridge") then
                cmd = "conference " .. conference_name .. " play /usr/local/freeswitch/prompt/unmute_allparty.wav " ..tostring(MemberID);
                freeswitch.consoleLog("notice", "[ConfAPI] : cmd : ".. tostring(cmd) .. "")
                response = api:executeString(cmd);
        end

