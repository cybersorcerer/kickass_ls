local log = require("plenary.log").new({
	plugin = "lsp.6510_ls",
	level = "info",
})
local capabilities = require("cmp_nvim_lsp").default_capabilities()

-- 6510 LSP Configuration Profiles
local config_profiles = {
	-- Default configuration - all features enabled
	default = {
		warnUnusedLabels = true,
		zeroPageOptimization = {
			enabled = true,
			showHints = true,
		},
		branchDistanceValidation = {
			enabled = true,
			showWarnings = true,
		},
		illegalOpcodeDetection = {
			enabled = true,
			showWarnings = true,
		},
		hardwareBugDetection = {
			enabled = true,
			showWarnings = true,
			jmpIndirectBug = true,
		},
		memoryLayoutAnalysis = {
			enabled = true,
			showIOAccess = true,
			showStackWarnings = true,
			showROMWriteWarnings = true,
		},
		magicNumberDetection = {
			enabled = true,
			showHints = true,
			c64Addresses = true,
		},
		deadCodeDetection = {
			enabled = true,
			showWarnings = true,
		},
		styleGuideEnforcement = {
			enabled = true,
			showHints = true,
			upperCaseConstants = true,
			descriptiveLabels = true,
		},
	},

	-- Legacy code profile - less strict
	legacy = {
		warnUnusedLabels = false,
		zeroPageOptimization = { enabled = true, showHints = false },
		branchDistanceValidation = { enabled = true, showWarnings = true },
		illegalOpcodeDetection = { enabled = false, showWarnings = false },
		hardwareBugDetection = { enabled = true, showWarnings = true, jmpIndirectBug = true },
		memoryLayoutAnalysis = { enabled = false },
		magicNumberDetection = { enabled = false },
		deadCodeDetection = { enabled = false },
		styleGuideEnforcement = { enabled = false },
	},

	-- Minimal profile - only critical errors
	minimal = {
		warnUnusedLabels = false,
		zeroPageOptimization = { enabled = false },
		branchDistanceValidation = { enabled = true, showWarnings = true },
		illegalOpcodeDetection = { enabled = false },
		hardwareBugDetection = { enabled = true, showWarnings = true, jmpIndirectBug = true },
		memoryLayoutAnalysis = { enabled = false },
		magicNumberDetection = { enabled = false },
		deadCodeDetection = { enabled = false },
		styleGuideEnforcement = { enabled = false },
	},
}

-- Get current profile (can be overridden by user)
local function get_current_profile()
	-- Check for project-specific configuration
	local project_config_file = vim.fn.findfile(".6510lsp.json", ".;")
	if project_config_file ~= "" then
		local ok, content = pcall(vim.fn.readfile, project_config_file)
		if ok then
			local profile_data = vim.fn.json_decode(table.concat(content, "\n"))
			if profile_data and profile_data["6510lsp"] then
				log.info("Using project-specific configuration from " .. project_config_file)
				return profile_data["6510lsp"]
			end
		end
	end

	-- Default to user preference or 'default' profile
	local user_profile = vim.g.c64_lsp_profile or "default"
	log.info("Using configuration profile: " .. user_profile)
	return config_profiles[user_profile] or config_profiles.default
end

-- LSP Settings Update Function
local function update_lsp_settings(new_settings)
	local clients = vim.lsp.get_clients({ name = "6510lsp" })
	for _, client in ipairs(clients) do
		client.notify("workspace/didChangeConfiguration", {
			settings = {
				["6510lsp"] = new_settings,
			},
		})
		log.info("Updated 6510 LSP configuration")
	end
end

-- User Commands for runtime configuration
vim.api.nvim_create_user_command("C64LspProfile", function(opts)
	local profile_name = opts.args
	if config_profiles[profile_name] then
		vim.g.c64_lsp_profile = profile_name
		update_lsp_settings(config_profiles[profile_name])
		print("Switched to profile: " .. profile_name)
	else
		print("Available profiles: " .. table.concat(vim.tbl_keys(config_profiles), ", "))
	end
end, {
	nargs = 1,
	complete = function()
		return vim.tbl_keys(config_profiles)
	end,
	desc = "Switch 6510 LSP configuration profile",
})

vim.api.nvim_create_user_command("C64LspToggleHints", function()
	local current_settings = get_current_profile()
	-- Toggle hints for various features
	current_settings.zeroPageOptimization.showHints = not current_settings.zeroPageOptimization.showHints
	current_settings.magicNumberDetection.showHints = not current_settings.magicNumberDetection.showHints
	current_settings.styleGuideEnforcement.showHints = not current_settings.styleGuideEnforcement.showHints

	update_lsp_settings(current_settings)
	print("Toggled 6510 LSP hints")
end, { desc = "Toggle 6510 LSP hints on/off" })

return {
	log.debug("Configuring 6510 Assembler LSP"),

	cmd = { "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server" },

	filetypes = { "6510", "asm", "c64asm" },

	root_markers = { ".6510lsp", ".git" },

	capabilities = capabilities,

	-- LSP Settings
	settings = {
		["6510lsp"] = get_current_profile(),
	},

	-- Enhanced on_attach callback
	on_attach = function(client, bufnr)
		log.info("6510 LSP attached to buffer " .. bufnr .. " (" .. vim.api.nvim_buf_get_name(bufnr) .. ")")

		-- Add 6510-specific keybindings
		local opts = { buffer = bufnr, noremap = true, silent = true }

		-- Toggle configuration on the fly
		vim.keymap.set("n", "<leader>ct", function()
			vim.cmd("C64LspToggleHints")
		end, vim.tbl_extend("force", opts, { desc = "Toggle C64 LSP hints" }))

		-- Show current configuration
		vim.keymap.set("n", "<leader>cs", function()
			local current_profile = vim.g.c64_lsp_profile or "default"
			print("Current 6510 LSP profile: " .. current_profile)
		end, vim.tbl_extend("force", opts, { desc = "Show C64 LSP status" }))
	end,
}
