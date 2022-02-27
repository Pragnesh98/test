--[[
  The Initial Developer of the Original Code is
  
  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  disconnect_all_line.lua --> conference controls 
--]]

	dofile("/usr/share/freeswitch/scripts/logger.lua");
	api = freeswitch.API();
	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
	xml = require "xml";
	
	freeswitch.consoleLog("notice", "[ConfAPI] :  Disconnect all lines except leader's - Leader Only");
	sounds_dir = session:getVariable("prompt_dir");
	conference_name = session:getVariable("conference_name");
	CallUUID = session:getVariable("uuid");
	
	cmd = "conference " .. conference_name .. " hup non_moderator"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
