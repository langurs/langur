// langur/ast/compiler_program.go

package ast

import (
	"fmt"
	"langur/common"
	"langur/opcode"
	"langur/str"
)

const moduleorder = "module/imports/modes/declarations"
const nonmoduleorder = "imports/expressions(including modes and declarations)"

func (c *Compiler) compileProgram(node *Program, executeModule bool) (
	pkg opcode.InsPackage, err error) {

	defer func() {
		if err == nil {
			err = c.checkStatementCounts()
		}
	}()

	if node.Search(nil, modeNotGlobal_searchCriteria) {
		err = c.makeErr(node, fmt.Sprintf("This implementation can only set modes in global context, or in the %s function", common.MainFnName))
		return
	}

	var mainStatementsStart int
	importsDone := false

	for i, s := range node.Statements {
		switch n := s.(type) {
		case *ModuleNode:
			if i == 0 {
				if n.Name != "" {
					err = c.makeErr(node, "No name expected on module")
					return
				}

				c.moduleDeclaredImpureEffects = n.ImpureEffects

				// A module has a more defined structure that must be followed.
				return c.compileModule(node.Statements[1:], executeModule)

			} else {
				// not first node; an error
				err = c.makeErr(node, "Module must be first part of code to compile as module")
				return
			}

		case *ImportNode:
			if importsDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", nonmoduleorder))
				return
			}

		case nil:
			err = c.makeErr(node, "Unexpected nil node")
			return

		default:
			// not a module or import node
			if !importsDone {
				mainStatementsStart = i
			}
			importsDone = true
		}
	}

	// wrap remaining instructions into main function and continue
	if importsDone {
		nodes := make([]Node, len(node.Statements[:mainStatementsStart])+1)
		copy(nodes, node.Statements[:mainStatementsStart])
		nodes[len(nodes)-1] = MakeMainFnDeclaration(node.Statements[mainStatementsStart:])

		// assume non-module level may contain impure effects, such as console or file changes
		c.moduleDeclaredImpureEffects = true
		return c.compileModule(nodes, true)
	}

	err = c.makeErr(node, "Expected statements/expressions")
	return
}

func (c *Compiler) compileModule(nodes []Node, execute bool) (
	pkg opcode.InsPackage, err error) {

	var modes []*ModeNode
	var modeNames []string
	var declarations []*DeclarationNode
	var imports []*ImportNode
	var bytes opcode.InsPackage

	importsDone := false
	modesDone := false

	for _, s := range nodes {
		switch node := s.(type) {
		case *ImportNode:
			if importsDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", moduleorder))
				return
			}
			imports = append(imports, node)

		case *ModeNode:
			if modesDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", moduleorder))
				return
			}
			importsDone = true

			if str.IsInSlice(node.Name, modeNames) {
				err = c.makeErr(node, fmt.Sprintf("Repeat of mode setting for %s", node.Name))
				return
			}
			modes = append(modes, node)
			modeNames = append(modeNames, node.Name)

		case *ExpressionStatementNode:
			decl, ok := node.Expression.(*DeclarationNode)
			if !ok {
				err = c.makeErr(node, fmt.Sprintf("Expected declarations only; cannot use other expressions in module context; use %s for a main function (if applicable)", common.MainFnName))
				return
			}
			importsDone = true
			modesDone = true

			// if possible, split up multi-variable assignments (or "flatten")
			var flatten []*DeclarationNode
			flatten, err = FlattenDeclaration(decl)
			if err != nil {
				return
			}

			declarations = append(declarations, flatten...)

		case nil:
			err = c.makeErr(node, "Unexpected nil node")
			return

		default:
			err = c.makeErr(node, fmt.Sprintf("Expected imports/modes/declarations, not %T", node))
			return
		}
	}

	// first compile import statements
	for _, importstmt := range imports {
		bytes, err = c.compileNodeWithPopIfExprStmt(importstmt)
		if err != nil {
			return
		}
		pkg = pkg.Append(bytes)
	}

	// then compile mode statements
	for _, mode := range modes {
		bytes, err = c.compileNodeWithPopIfExprStmt(mode)
		if err != nil {
			return
		}
		pkg = pkg.Append(bytes)
	}

	// last of all, compile declarations
	declarations, err = c.fixModuleDeclarations(declarations)
	if err != nil {
		return
	}
	for _, decl := range declarations {
		bytes, err = c.compileNodeWithPopIfExprStmt(decl)
		if err != nil {
			return
		}
		pkg = pkg.Append(bytes)
	}

	if execute {
		bytes, err = c.compileNodeWithPopIfExprStmt(ExecuteMain)
		if err != nil {
			return
		}
		pkg = pkg.Append(bytes)
	}

	if c.impureEffects && !c.moduleDeclaredImpureEffects {
		err = c.makeErr(nodes[0], "Module contains impure effects and is not declared impure; use module* (with asterisk)")
	}

	return
}

func (c *Compiler) fixModuleDeclarations(declarations []*DeclarationNode) (
	decl []*DeclarationNode, err error) {

	L := len(declarations)
	decl = make([]*DeclarationNode, L)

	for i := range declarations {
		if declarations[i].Mutable {
			// disallow module var declarations for now
			// Within a function, the variables would be closures.
			// Since we can't mutate them anyway, allowing mutable declarations is confusing.
			return nil, c.makeErr(declarations[i], "Cannot use var declarations in module context in this implementation/version")
		}

		// fix function declarations that won't pass as system functions
		a, ok := declarations[i].Assignment.(*AssignmentNode)
		if ok {
			if len(a.Identifiers) == 1 {
				id, ok := a.Identifiers[0].(*IdentNode)
				if ok {
					switch id.Name {
					case common.MainFnName:
						// set to system to make it work
						declarations[i].Assignment.(*AssignmentNode).Identifiers[0].(*IdentNode).System = true

						// verify not mutable and is a function
						if declarations[i].Mutable {
							return decl, c.makeErr(declarations[i], fmt.Sprintf("%s must be immutable declaration", id.Name))
						}
						_, isFunction := a.Values[0].(*FunctionNode)
						if !isFunction {
							return decl, c.makeErr(declarations[i], fmt.Sprintf("%s must be a function", id.Name))
						}
					}
				}
			}
		}

		// reverse order to put dependent functions at top
		decl[L-i-1] = declarations[i]
	}

	return decl, nil
}
