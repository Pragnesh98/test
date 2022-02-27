--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  moderator_control.lua --> conference controls
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

        function play_member_count(announce_count, member_count, member_count_prompt)
                --play member count
                if (announce_count == "true") then
			--if (member_count == "2") then
                        --there is one other member in this conference
                          --      session:execute("playback", "/usr/local/freeswitch/prompt/there_is_one.wav");
                        if (member_count == "1") then
--                              if (wait_mod == "true" and member_type ~= "moderator") then
                                session:execute("playback", "/usr/local/freeswitch/prompt/there_is_one.wav");
                                      --  session:execute("playback", member_count_prompt);
--                              end
                                --conference profile defines the alone sound file
                        else
                                session:execute("playback", "/usr/local/freeswitch/prompt/there-are.wav");
                                --say the count
                                session:execute("say", "en number pronounced "..member_count);
                                --members in this conference
                                session:execute("playback", "/usr/local/freeswitch/prompt/party_conf.wav");
                        end

	end
end

local function Counts(session, conference_name, CallUUID)
        conference_name = session:getVariable("conference_name");
        conference_name = argv[1];
        CallUUID = session:getVariable("uuid");
	CallUUID = argv[2]
        sounds_dir = session:getVariable("prompt_dir");
	sounds_dir =freeswitch.getGlobalVariable("prompt_dir")

	cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

        cmd = "conference "..tostring(conference_name).." list count";
        member_count = api:executeString(cmd);
        if string.match(member_count, "not found") then
                member_count = "0";
        end
        freeswitch.consoleLog("notice", "[ConfAPI] :  member_count : "..tostring(member_count));
        play_member_count("true", member_count, sounds_dir.."/451.wav")
end

return Counts
