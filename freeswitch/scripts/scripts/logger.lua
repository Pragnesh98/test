--[[
	The Initial Developer of the Original Code is
	XYZ pvt ltd
	Portions created by the Initial Developer are Copyright (C)
	the Initial Developer. All Rights Reserved.

	Contributor(s):

	logger.lua : write a Log in FS console
--]]

	Logger = {}
	Logger.__index = Logger

	function Logger.console(message) 
	Logger.print("console",message)
	end

	function Logger.alert(message)  
	Logger.print("alert",message)
	end

	function Logger.critical(message)  
	Logger.print("critical",message)
	end

	function Logger.error(message)  
	Logger.print("err",message)
	end

	function Logger.warning(message)  
	Logger.print("warning",message)
	end

	function Logger.notice(message)  
	Logger.print("notice",message)
	end

	function Logger.info(message)  
	Logger.print("info",message)
	end

	function Logger.debug(message)  
	Logger.print("debug",message)
	end

	function Logger.print(logtype,message)
		if(message ~= nil) then
			freeswitch.consoleLog(logtype,"[Logger] "..message.. "\n");
		end
	end
