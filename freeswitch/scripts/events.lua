--[[
        The Initial Developer of the Original Code is
        ucall <https://uvoice.ucall.co.ao/> [ucall]

        Portions created by the Initial Developer are Copyright (C)
        the Initial Developer. All Rights Reserved.

        Contributor(s):
        ucall <https://uvoice.ucall.co.ao/> [ucall]
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

--      package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
--      package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        xml = require "xml";


	--require "resources.functions.config";
	debug["sql"] = false;
	local json
	if (debug["sql"]) then
	json = require "resources.functions.lunajson"
	end

	--local Database = require "resources.functions.database";
	--dbh = Database.new('system');
	--assert(dbh:connected());

	dest_queue_uuid = "";
	api = freeswitch.API();
	local event_name = event:getHeader("Event-Name");

	UrL = freeswitch.getGlobalVariable("UrL")

	if event_name == "CHANNEL_HANGUP" then
		freeswitch.consoleLog("notice","[events] event_name : ["..event_name.."]\n");
		uuid = event:getHeader("Unique-ID");
                freeswitch.consoleLog("notice", "[events] uuid : ["..tostring(uuid).."]\n");

		conference_name = event:getHeader("variable_conference_name");
                freeswitch.consoleLog("notice", "[events] Conference_Name : ["..tostring(conference_name)..", "..script_dir.."]\n");

		member_type = event:getHeader("variable_member_type");
         --       freeswitch.consoleLog("notice", "[events] Conference_Name : ["..tostring(member_type).."]\n");
		
		conference_continuation = event:getHeader("variable_conference_continuation");
       --         freeswitch.consoleLog("notice", "[events] Conference_continuation : ["..tostring(conference_continuation).."]\n");
		
--[[		if (member_type == "moderator" and (conference_continuation == "notallow" or conference_continuation == nil) ) then
                        
			cmd="conference "..tostring(conference_name).." kick all"
			Logger.notice("[MindBridge] : API : "..tostring(cmd));
                        response =  api:executeString(cmd);
			Logger.notice("[MindBridge] : response  : "..tostring(response));

		end --]]
   --[[       		cmd="conference "..tostring(conference_name).." play /usr/local/freeswitch/prompt/now-exiting.wav"
			Logger.notice("[MindBridge] : API : "..tostring(cmd));
                        response =  api:executeString(cmd);

		cmd = "conference "..tostring(conference_name).." play /tmp/conference-"..uuid..".mp3";
		--freeswitch.consoleLog("notice","[events]  cmd : "..cmd.."\n");
                response = api:executeString(cmd);
		freeswitch.consoleLog("notice","[events]  response : "..response.."\n");
]]--

		cmd = "conference "..tostring(conference_name).." list count";
                freeswitch.consoleLog("notice","[MindBridge] : API : "..tostring(cmd).."");
                member_count = api:executeString(cmd);
                freeswitch.consoleLog("notice","[MindBridge] : Member_count : "..tostring(member_count).."");
                if tonumber(member_count) == 1 then
                        cmd="curl --location --request GET 'http://localhost:10000/delete-queue'"
                        freeswitch.consoleLog("notice","[MindBridge] : API : "..tostring(cmd).."");
                        curl_response = api:execute("system", cmd);
                        freeswitch.consoleLog("notice","[MindBridge] :curl_response : "..tostring(curl_response).."");

			APIcmd = "curl --location --request DELETE '"..tostring(UrL).."delete-conference-server-ip/"..tostring(conference_name).."'"
                        freeswitch.consoleLog("notice", "[MindBridge] : Conf Server DELETE Request : "..tostring(APIcmd).."");
                        curl_response = api:execute("system", APIcmd);
                        freeswitch.consoleLog("notice","[MindBridge] : Conf Server DELETE Response : "..tostring(curl_response).."");
                        freeswitch.consoleLog("notice","[MindBridge] : Delete Conference From DB");

                end


		cmd="curl --location --request GET 'http://localhost:10000/queue/"..tostring(uuid).."/"..tostring(conference_name).."'"
                freeswitch.consoleLog("notice","[MindBridge] : API : "..tostring(cmd).."");
                curl_response = api:execute("system", cmd);
                freeswitch.consoleLog("notice","[MindBridge] :curl_response : "..tostring(curl_response).."");




         --[[     		cmd="conference "..tostring(conference_name).." moh off"
			Logger.notice("[MindBridge] : API : "..tostring(cmd));
                        response =  api:executeString(cmd);
			Logger.notice("[MindBridge] : Response : "..tostring(response));
	]]--	
			serialized = event:serialize('json');
	--	freeswitch.consoleLog("notice", "[events] Event : ["..serialized.."]\n");
	end
