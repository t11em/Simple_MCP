package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/t11em/Simple_MCP"
)

type RollDiceArguments struct {
	Sides int `json:"sides"`
}

func registerRollDiceTool(h *simplemcp.Handler) {
	h.RegisterTool(&simplemcp.RegisterToolConfig{
		Name:        "roll_dice",
		Description: "Roll a dice",
		Properties: map[string]simplemcp.Property{
			"sides": {
				Type:        simplemcp.PropertyTypeInteger,
				Description: "Number of sides on the dice.",
			},
		},
		// Note: 戻り値を指定しないとエラーになったため修正
		ToolFunc: func(ctx context.Context, args json.RawMessage) (simplemcp.CallToolResult, error) {
			var (
				rollDiceArgs = RollDiceArguments{}
				result       = simplemcp.NewCallToolResult()
			)
			err := json.Unmarshal(args, &rollDiceArgs)
			if err != nil {
				result.IsError = true
				result.AddTextContent("Invalid parameters")
				return result, err
			}
			diceResult := rand.Intn(rollDiceArgs.Sides) + 1
			resultTxt := strconv.Itoa(diceResult)
			if rollDiceArgs.Sides == 100 {
				if diceResult <= 5 {
					// TRPGの文脈？
					resultTxt = fmt.Sprintf("%d:クリティカル", diceResult)
				} else if diceResult <= 20 {
					resultTxt = fmt.Sprintf("%d:成功", diceResult)
				} else if diceResult <= 95 {
					resultTxt = fmt.Sprintf("%d:失敗", diceResult)
				} else {
					// TRPGの文脈？
					resultTxt = fmt.Sprintf("%d:ファンブル", diceResult)
				}
			}
			result.AddTextContent(resultTxt)
			return result, nil
		},
	})
}

func main() {
	h := simplemcp.NewHandler(&simplemcp.Implementation{
		Name:    "dice_roller",
		Version: "0.0.1",
	})
	registerRollDiceTool(h)
	h.Run(context.Background())
}
