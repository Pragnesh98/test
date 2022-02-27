--[[
  The Initial Developer of the Original Code is
  
  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  start_recording.lua --> conference centers 
--]]
	conference_name = argv[1];
	uuid = argv[2];

	dofile("/usr/share/freeswitch/scripts/logger.lua");

--general functions
	require "resources.functions.base64";
	require "resources.functions.trim";
	require "resources.functions.file_exists";
	require "resources.functions.explode";
	require "resources.functions.format_seconds";
	require "resources.functions.mkdir";
	
	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
	xml = require "xml";
	
--prepare the api object
	api = freeswitch.API();

--check if the conference exists
	cmd = "conference "..conference_name.." xml_list";
	freeswitch.consoleLog("INFO","" .. cmd .. "\n");
	result = api:executeString(cmd)
	if string.match(result, "not found") then
		conference_exists = false;
	else
		conference_exists = true;
	end

--start the recording
	if (conference_exists) then
		--get the conference session uuid
			result = string.match(result,[[<conference (.-)>]],1);
			conference_session_uuid = string.match(result,[[uuid="(.-)"]],1);
			freeswitch.consoleLog("INFO","[start-recording] conference_session_uuid: " .. conference_session_uuid .. "\n");

		--get the current time
			start_epoch = os.time();

		--add the domain name to the recordings directory
			recordings_dir = "/var/lib/freeswitch/recordings";
			recordings_dir = recordings_dir.."/archive/"..os.date("%Y", start_epoch).."/"..os.date("%b", start_epoch).."/"..os.date("%d", start_epoch);
			mkdir(recordings_dir);
			recording = recordings_dir.."/"..conference_session_uuid..".wav";
			
			cmd = "uuid_setvar "..tostring(uuid).." conference_call_recording "..recording;
			freeswitch.consoleLog("notice", "[start-recording] cmd: " .. cmd .. "\n");
			response = api:executeString(cmd);
				
		--send a command to record the conference
			if (not file_exists(recording)) then
				cmd = "conference "..conference_name.." record "..recording;
				freeswitch.consoleLog("notice", "[start-recording] cmd: " .. cmd .. "\n");
				response = api:executeString(cmd);
			end
	end
