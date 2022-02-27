--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  start_recording.lua --> conference centers
--]]
        conference_name = argv[1];
        uuid = argv[2];

        script_dir = freeswitch.getGlobalVariable("script_dir")
        dofile(script_dir.."/logger.lua");
        if(script_dir == "/usr/local/freeswitch/scripts") then
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/local/share/lua/5.1/?.lua;" .. package.path
        else
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        end
        xml = require "xml";

--prepare the api object
        api = freeswitch.API();
--check if the conference exists
	--conference_name = session:getVariable("conference_name");
        freeswitch.consoleLog("NOTICE","" .. conference_name .. "\n");
        --cmd = "conference "..conference_name.." xml_list";
        --freeswitch.consoleLog("INFO","" .. cmd .. "\n");
        --result = api:executeString(cmd)
        --if string.match(result, "not found") then
          --      conference_exists = false;
        --else
          --      conference_exists = true;
        --end

		--if(conference_exists) then
			cmd = "conference "..conference_name.." moh off";
			freeswitch.consoleLog("notice", "[moh] cmd: " .. cmd .. "\n");
			response = api:executeString(cmd);
		--end

