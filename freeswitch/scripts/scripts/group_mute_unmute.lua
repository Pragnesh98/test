--[[
  The Initial Developer of the Original Code is
  
  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  lecture_mode.lua --> conference controls 
--]]

	dofile("/usr/share/freeswitch/scripts/logger.lua");
	api = freeswitch.API();
	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
	xml = require "xml";
	
	freeswitch.consoleLog("notice", "[ConfAPI] :  Group Mute/Unmute");
	conference_name = session:getVariable("conference_name");
	CallUUID = session:getVariable("uuid");

	group_mute_state = session:getVariable("group_mute_state");
	if (group_mute_state == "true" ) then
		session:setVariable("group_mute_state","false")
		conference_name = session:getVariable("conference_name");
		cmd = "conference " .. conference_name .. " unmute all" ;
		response = api:executeString(cmd);
			
		freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
		freeswitch.consoleLog("notice", "[ConfAPI] :  Group unmuted");
	else
		session:setVariable("group_mute_state","true")
		conference_name = session:getVariable("conference_name");
		cmd = "conference " .. conference_name .. " mute all";
		response = api:executeString(cmd);
			
		freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
		freeswitch.consoleLog("notice", "[ConfAPI] :  Group muted");
	end
