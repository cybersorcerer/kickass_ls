-- lsp configuration as of neovim 0.11
vim.api.nvim_create_autocmd("LspAttach", {
	desc = "Configure Lsp's on attach",
	group = vim.api.nvim_create_augroup("lsp_configure", { clear = true }),
	callback = function(ev)
		local client = vim.lsp.get_client_by_id(ev.data.client_id)
		local bufmap = function(mode, rhs, lhs)
			vim.keymap.set(mode, rhs, lhs, { buffer = ev.buf })
		end

		-- These keymaps are the defaults in Neovim v0.11
		bufmap("n", "K", "<cmd>lua vim.lsp.buf.hover()<cr>")
		bufmap("n", "grr", "<cmd>lua vim.lsp.buf.references()<cr>")
		bufmap("n", "gri", "<cmd>lua vim.lsp.buf.implementation()<cr>")
		bufmap("n", "grn", "<cmd>lua vim.lsp.buf.rename()<cr>")
		bufmap("n", "gra", "<cmd>lua vim.lsp.buf.code_action()<cr>")
		bufmap("n", "gO", "<cmd>lua vim.lsp.buf.document_symbol()<cr>")
		bufmap({ "i", "s" }, "<C-s>", "<cmd>lua vim.lsp.buf.signature_help()<cr>")

		-- These are custom keymaps
		bufmap("n", "gd", "<cmd>lua vim.lsp.buf.definition()<cr>")
		bufmap("n", "grt", "<cmd>lua vim.lsp.buf.type_definition()<cr>")
		bufmap("n", "grd", "<cmd>lua vim.lsp.buf.declaration()<cr>")
		bufmap({ "n", "x" }, "gq", "<cmd>lua vim.lsp.buf.format({async = true})<cr>")

		-- 6510 Assembly specific keymaps
		if client and client.name == "6510lsp" then
			-- Toggle configuration on the fly
			bufmap("n", "<leader>ct", function()
				vim.cmd("C64LspToggleHints")
			end, { desc = "Toggle C64 LSP hints" })

			-- Show current configuration
			bufmap("n", "<leader>cs", function()
				local current_profile = vim.g.c64_lsp_profile or "default"
				print("Current 6510 LSP profile: " .. current_profile)
			end, { desc = "Show C64 LSP status" })
		end

		--if client:supports_method("textDocument/completion") then
		--	vim.opt.completeopt = { "menu", "menuone", "noinsert", "fuzzy", "popup" }
		--	vim.lsp.completion.enable(true, client.id, ev.buf, { autotrigger = true })
		--	vim.keymap.set("i", "<C-Space>", function()
		--		vim.lsp.completion.get()
		--	end)
		--end
	end,
})

-- Diagnostics configuration
vim.diagnostic.config({
    virtual_lines = {
        current_line = true,
    },
})

-- 6510 Assembly LSP Integration for ronny.core structure
local M = {}

-- Function to load 6510 LSP configuration from ~/.config/nvim/lsp/6510_ls.lua
function M.setup_6510_lsp()
	local config_file = vim.fn.stdpath("config") .. "/lsp/6510_ls.lua"

	-- Check if 6510_ls.lua exists in the lsp directory
	if vim.fn.filereadable(config_file) ~= 1 then
		return -- Silently skip if config doesn't exist
	end

	-- Load 6510 LSP configuration
	local ok, config = pcall(function()
		return dofile(config_file)
	end)

	if not ok or not config then
		vim.notify("Failed to load 6510 LSP configuration from " .. config_file, vim.log.levels.ERROR)
		return
	end

	-- Auto-start LSP for 6510 assembly files
	vim.api.nvim_create_autocmd({"BufEnter", "BufWinEnter"}, {
		pattern = {"*.asm", "*.s", "*.6510", "*.inc", "*.a"},
		group = vim.api.nvim_create_augroup("6510_lsp_start", { clear = true }),
		callback = function()
			local clients = vim.lsp.get_clients({ name = "6510lsp" })
			if #clients == 0 then
				-- Start the LSP with configuration from ~/.config/nvim/lsp/6510_ls.lua
				vim.lsp.start(vim.tbl_extend("force", config.lsp_config, {
					name = "6510lsp",
					root_dir = vim.fs.root(0, config.lsp_config.root_markers) or vim.fn.getcwd(),
				}))
			end
		end,
	})

	-- Additional 6510-specific commands
	vim.api.nvim_create_user_command("C64LspRestart", function()
		local clients = vim.lsp.get_clients({ name = "6510lsp" })
		for _, client in ipairs(clients) do
			client.stop()
		end
		vim.defer_fn(function()
			M.setup_6510_lsp()
			print("6510 LSP restarted")
		end, 500)
	end, { desc = "Restart 6510 LSP server" })
end

-- Initialize 6510 LSP on module load
M.setup_6510_lsp()

return M
