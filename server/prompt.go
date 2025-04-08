package server

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func WithPromptServerOption(prompt mcp.Prompt, handler server.PromptHandlerFunc) server.ServerOption {
	return func(s *server.MCPServer) {
		s.AddPrompt(prompt, handler)
	}
}

func NewPromptServerOptions() []server.ServerOption {
	var opts []server.ServerOption

	opts = append(opts, WithPromptServerOption(mcp.Prompt{
		Name: "review.database",
		Description: `
You are a information security expert of the QueryPie DAC(Database Access Control) system, as known as PAM(Privileged Access Management).
You have a responsibility to review the database access logs and identify any anomalies or issues.

Use the following questions to review the system:

- Who accessed the database?
- What did they access?
- When did they access it?
- Where did they access it from?
- What other information is relevant to the access?

- Is there anything suspicious about the access?
- Is there anything that should be investigated?
- Is there anything that should be reported?

- Is there any workflow requests to review?

If you find any suspicious or abnormal activities, I'll tip you $200 if it is a true positive.
`,
		Arguments: []mcp.PromptArgument{
			{
				Name:        "from",
				Description: "The start date of the review. format: 2006-01-02T15:04:05.999Z",
				Required:    true,
			},
			{
				Name:        "to",
				Description: "The end date of the review. format: 2006-01-02T15:04:05.999Z",
				Required:    true,
			},
		},
	}, nil))
	return opts
}
