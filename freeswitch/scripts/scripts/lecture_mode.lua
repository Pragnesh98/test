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
	
	Logger.notice ("[ConfAPI] : conference Lecture mode enabled")
	conference_name = session:getVariable("conference_name");
	CallUUID = session:getVariable("uuid");
	
	lecture_mode = session:getVariable("lecture_mode");
	if (lecture_mode == "true" ) then
		session:setVariable("lecture_mode","false")
		conference_name = session:getVariable("conference_name");
		cmd = "conference " .. conference_name .. " unmute all" ;
		response = api:executeString(cmd);
		freeswitch.consoleLog("notice", "[ConfAPI] :  Lecture mode Disabled API : "..tostring(cmd));
	else
		session:setVariable("lecture_mode","true")
		conference_name = session:getVariable("conference_name");
		cmd = "conference " .. conference_name .. " mute all";
		response = api:executeString(cmd);
		freeswitch.consoleLog("notice", "[ConfAPI] :  Lecture mode Enabled API : "..tostring(cmd));
	end

	-- while unmute need to check chairperson lecture_mode and base on that need to allow unmute.
