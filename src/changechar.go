package main

// ChangeChar state controller (Ikemen-only).
// Replaces a character with a different .def file, loading the entire
// character (.def/.sff/.air/.snd/.cmd/.st), reusing the same reload
// mechanism as MatchRestart's pXdef but limited to a single target.
//
// Parameters:
//
//	value    = "char_path" (string) Path of the .def file to load.
//	           Relative to the folder of the character that triggered
//	           ChangeChar or the top ikemen directory.
//	playerid = id (int) Target character's PlayerID. If omitted, the
//	           character that triggered ChangeChar is replaced.
//
// The target character's variables are preserved across the swap.
type changeChar StateControllerBase

const (
	changeChar_playerid byte = iota
	changeChar_value
)

func (sc changeChar) Run(c *Char, _ []int32) bool {
	dirs := []string{c.gi().def, "", "data/"}
	var value string
	playerid := int32(-1)

	StateControllerBase(sc).run(c, func(paramID byte, exp []BytecodeExp) bool {
		switch paramID {
		case changeChar_playerid:
			playerid = exp[0].evalI(c)
		case changeChar_value:
			value = exp[0].evalS(c)
		}
		return true
	})

	if value == "" {
		return false
	}

	// Resolve the target character: explicit playerid or the owner itself.
	target := c
	if playerid >= 0 {
		if t := sys.playerID(playerid); t != nil {
			target = t
		} else {
			return false
		}
	}

	// Queue the swap. Preserve the character's variables across the reload.
	sys.sel.cdefOverwrite[target.playerNo] = SearchFile(value, dirs, "chars/")
	sys.reloadCharSlot[target.playerNo] = true
	sys.reloadPreserveVars[target.playerNo] = true
	if sys.netConnection == nil && sys.replayFile == nil {
		sys.reloadFlg = true
	}
	return false
}

func (c *CharCompiler) changeChar(is IniSection, sc *StateControllerBase) (StateController, error) {
	ret, err := (*changeChar)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "value",
			changeChar_value, VT_String, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerid",
			changeChar_playerid, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
