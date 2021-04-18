package sqlhelpers

import (
	"fmt"
	"strings"
)

const (
	DeleteField MigrateType = iota
	Rename
	SetDefault
	SetPrimaryKey
	SetReference
	SetNextVal
	Add
	Remove
	Update
)

//Migrate creates analyzes two structs of the same type and will return various sql statements to migrate from old to new.
//If old is nil it returns sql for a new table.
//If passed true in the last argument `safe`, it will not return drop column statements.
func Migrate(oldMap Table, new interface{}, tableName string, safe bool) (res []string, newMap Table, err error) {
	if oldMap == nil {
		return []string{CreateTable(new, tableName)}, nil, nil
	}
	newMap, err = StructMap(new)
	if err != nil {
		return nil, nil, err
	}
	queries := []string{}
	for key, oldval := range oldMap {
		newval, ok := newMap[key]
		switch {
		case !ok:
			//doesnt exist in new
			if !safe {
				queries = append(queries, fmt.Sprintf("alter table %v drop column %q", tableName, oldval.JSON))
			}
		case ok:
			//exists in both
			if oldval.JSON != newval.JSON { // name changed

				oldIsField, _ := isField(oldval.DB)
				newIsField, _ := isField(newval.DB)
				if oldIsField && newIsField {
					//rename
					queries = append(queries, fmt.Sprintf("alter table %v rename column %q to %q", tableName, oldval.JSON, newval.JSON))
				}
			}
			if oldval.DB != newval.DB {
				// fmt.Println("parsing db changes for", key)
				qs, err := parseMigration(tableName, newval.JSON, oldval.DB, newval.DB, safe)
				if err != nil {
					return nil, nil, err
				}
				queries = append(queries, qs...)
			}
		}
	}
	for key, newval := range newMap {
		if _, ok := oldMap[key]; !ok {
			ok, field := isField(newval.DB)
			if ok {
				queries = append(queries, fmt.Sprintf("alter table %v add column %q %v", tableName, newval.JSON, field))
			}
		}
	}
	return queries, newMap, nil
}

type MigrateType int

func parseMigration(table, name, old, new string, safe bool) ([]string, error) {
	oldIsField, oldField := isField(old)
	newIsField, newField := isField(new)
	// fmt.Printf("\told %q new %q\n", old, new)
	if oldIsField && !newIsField {
		//delete column
		if !safe {
			return []string{fmt.Sprintf("alter table %v drop column %q", table, name)}, nil
		}
		return nil, nil
	}
	if !oldIsField && newIsField {
		fmt.Printf("\tadd column %q %v\n", name, newField)
		return []string{fmt.Sprintf("alter table %v add column %q %v", table, name, newField)}, nil
		//add column
	}
	if oldIsField && newIsField {
		return parseChanges(table, name, oldField, newField)
		//parse changes
	}

	// fmt.Println("\t", oldIsField, newIsField)
	return nil, nil
}

var migrationTokens = map[string]func(string, string, []string, []string, MigrateType) ([]string, error){
	"primary key": parsePrimaryKey,
	"unique":      parseUnique,
	"not null":    parseNotNull,
	"default(":    parseDefault,
	"references":  parseReferences,
	"check(":      parseCheck,
}

func parseChanges(table, name, old, new string) ([]string, error) {
	statements := []string{}
	oldType := getType(old)
	newType := getType(new)
	if oldType != newType {
		sql := fmt.Sprintf("alter table %v alter column %v type %v using %v::%v", table, name, newType, name, newType)
		statements = append(statements, sql)
	}
	oldTokens := strings.Split(old, " ")
	newTokens := strings.Split(new, " ")
	for tk, fn := range migrationTokens {
		oldToken := strings.Index(old, tk)
		newToken := strings.Index(new, tk)
		if oldToken == -1 && newToken == -1 {
			//doesn't exist so skip
			continue
		}
		if oldToken == -1 && newToken > -1 {
			//it has been added or modified so create it
			res, err := fn(table, name, oldTokens, newTokens, Add)
			if err != nil {
				return nil, err
			}
			statements = append(statements, res...)
			continue
		}
		if oldToken > -1 && newToken == -1 {
			//it has been removed so drop it
			res, err := fn(table, name, oldTokens, newTokens, Remove)
			if err != nil {
				return nil, err
			}
			statements = append(statements, res...)
			continue
		}
		if oldToken > -1 && newToken > -1 {
			//both have it, check if it's different
			res, err := fn(table, name, oldTokens, newTokens, Update)
			if err != nil {
				return nil, err
			}
			statements = append(statements, res...)
			continue
		}

	}
	//alter column
	return statements, nil
}

func getType(s string) string {
	return strings.Split(s, " ")[0]
}
func isField(s string) (bool, string) {
	if ok := strings.Contains(s, "field:"); ok {
		if split := strings.Split(s, "field:"); len(split) > 1 {
			return ok, split[1]
		}
	}
	return false, ""
}

func parsePrimaryKey(table, name string, oldTokens, newTokens []string, typ MigrateType) ([]string, error) {
	switch typ {
	case Add:
		return []string{fmt.Sprintf("alter table %v add primary key (%q)", table, name)}, nil
	case Remove:
		return []string{fmt.Sprintf("alter table %v drop constraint if exists %v_pk", table, table)}, nil
	}
	return nil, nil
}
func parseNotNull(table, name string, oldTokens, newTokens []string, typ MigrateType) ([]string, error) {
	switch typ {
	case Add:
		return []string{fmt.Sprintf(`alter table %v alter column %q set not null`, table, name)}, nil
	case Remove:
		return []string{fmt.Sprintf(`alter table %v alter column %q drop not null`, table, name)}, nil
	}
	return nil, nil
}
func parseDefault(table, name string, oldTokens, newTokens []string, typ MigrateType) ([]string, error) {
	switch typ {
	case Add, Update:
		var df string
		for _, token := range newTokens {
			if strings.Contains(token, "default") {
				split := strings.Split(token, "default(")
				if len(split) > 1 {
					s := split[1]
					df = s[0 : len(s)-1]
				}
				break
			}
		}
		if df != "" {
			return []string{fmt.Sprintf(`alter table %v alter column %q set default %v`, table, name, df)}, nil
		}
		return nil, nil
	case Remove:
		return []string{fmt.Sprintf(`alter table %v alter column %q drop default`, table, name)}, nil
	}
	return nil, nil
}

var referenceActions = map[string]struct{}{
	"restrict":  empty,
	"cascade":   empty,
	"no action": empty,
	"set":       empty,
	"null":      empty,
	"default":   empty,
}

func parseReferences(table, name string, oldTokens, newTokens []string, typ MigrateType) ([]string, error) {
	//structure is = references <ident>(<ident>) <referenceToken> <referenceAction> <referenceToken> <referenceAction>
	res := []string{}

	if typ == Remove || typ == Update {
		res = append(res, fmt.Sprintf("alter table %v drop constraint if exists %v_%v_fk", table, table, name))
	}
	oldfield, oldreference, oldDel, oldUpd, err := parseReferenceTokens(table, name, oldTokens)
	if err != nil {
		return nil, err
	}
	field, reference, delAction, updAction, err := parseReferenceTokens(table, name, newTokens)
	if err != nil {
		return nil, err
	}
	if field == oldfield && reference == oldreference && oldDel == delAction && oldUpd == updAction {
		return nil, nil
	}

	if field == "" || reference == "" {
		return nil, fmt.Errorf("missing reference field or reference for %q column %q", table, name)
	}

	res = append(res, fmt.Sprintf("alter table %v add constraint %v_%v_fk foreign key (%q) references %v (%q) on update %v on delete %v", table, table, name, name, reference, field, updAction, delAction))
	// fmt.Sprintf("references %v(%v)  %v", val, "id", "on delete cascade on update cascade")
	return res, nil
}

func parseReferenceTokens(table, name string, tokens []string) (field string, reference string, delAction string, updAction string, err error) {
	ident := false
	del := false
	upd := false
	on := false
	set := false
	delAction = "no action"
	updAction = "no action"
	for _, token := range tokens {
		if ident {
			s := strings.SplitN(token, "(", 2)
			if len(s) > 1 {
				reference = s[0]
				field = s[1][0 : len(s[1])-1]
				ident = false
				on = false
				continue
			}
			break
		}
		if on {
			if token == "update" {
				upd = true
				del = false
				on = false
				continue
			}
			if token == "delete" {
				del = true
				upd = false
				on = false
				continue
			}
		}
		if del {
			if _, ok := referenceActions[token]; ok {
				if !set {
					if token == "set" {
						set = true
						continue
					}
					delAction = token
					del = false
					continue
				}
				if token != "null" && token != "default" {
					return "", "", "", "", fmt.Errorf("syntax error at migration for %q column %q. Expecting on delete set null or set default, got set %v", table, name, token)
				}
				delAction = "set " + token
				del = false
				continue
			}
			return "", "", "", "", fmt.Errorf("syntax error at migration for %q column %q. Unknown 'on delete' action: got %v, expected one of (cascade, set null, set default, no action, restrict) ", table, name, token)
		}
		if upd {
			if _, ok := referenceActions[token]; ok {
				if !set {
					if token == "set" {
						set = true
						continue
					}
					updAction = token
					upd = false
					continue
				}
				if token != "null" && token != "default" {
					return "", "", "", "", fmt.Errorf("syntax error at migration for %q column %q. Expecting on update set null or set default, got set %v", table, name, token)
				}
				updAction = "set " + token
				upd = false
				continue
			}
			return "", "", "", "", fmt.Errorf("syntax error at migration for %q column %q. Unknown 'on update' action: got %v, expected one of (cascade, set null, set default, no action, restrict) ", table, name, token)
		}
		switch token {
		case "references":
			ident = true
			continue
		case "on":
			on = true
		default:
			break
		}
	}
	return field, reference, delAction, updAction, nil
}

func parseCheck(table, name string, oldTokens, newTokens []string, typ MigrateType) ([]string, error) {
	res := []string{}
	if typ == Remove || typ == Update {
		res = append(res, fmt.Sprintf("alter table %v drop constraint if exists %v_%v_check", table, table, name))
	}
	if typ == Update || typ == Add {
		str := strings.Join(newTokens, " ")
		start := strings.Index(str, "check(")
		if start > -1 {
			var end int
			openpar := 0
		loop:
			for i := start; i < len(str); i++ {
				b := str[i]
				switch b {
				case '(':
					openpar++
				case ')':
					if openpar < 2 {
						end = i
						break loop
					}
					openpar--
				}
			}
			if start < end && end < len(str) {
				check := str[start : end+1]
				// fmt.Println(check)
				res = append(res, fmt.Sprintf("alter table %v add constraint %v_%v_check %v", table, table, name, check))
			}
		}
	}

	return res, nil
}
func parseUnique(table, name string, oldTokens, newTokens []string, typ MigrateType) ([]string, error) {
	res := []string{}
	if typ == Remove || typ == Update {
		res = append(res, fmt.Sprintf(`alter table %v drop constraint if exists %v_%v_key`, table, table, name))
	}
	if typ == Add || typ == Update {
		res = append(res, fmt.Sprintf(`alter table %v add constraint %v_%v_key unique (%v)`, table, table, name, name))
	}
	//ALTER TABLE public.users DROP CONSTRAINT users_email_key;

	// ALTER TABLE public.users ADD CONSTRAINT users_email_key UNIQUE (email)

	return res, nil
}

/*
default(1) check(plan = 1 or plan = 2 or plan = 3 or plan = 4)
references users(id) on delete cascade on update cascade
*/
