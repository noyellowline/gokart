package config

import (
    "fmt"
    "os"
    "regexp"
    "strings"
)

var (
    templateRegex = regexp.MustCompile(`\${([A-Za-z_][A-Za-z0-9_]*)(?::([^}]*))?}`)
    maxDepth      = 10
)

func expandTemplates(yamlContent []byte) ([]byte, []string, error) {
    content := string(yamlContent)
    expanded, unexpanded, err := expandRecursive(content, 0)
    if err != nil {
        return nil, nil, err
    }
    return []byte(expanded), unexpanded, nil
}

func expandRecursive(input string, depth int) (string, []string, error) {
    if depth > maxDepth {
        return "", nil, fmt.Errorf("template expansion exceeded maximum depth (%d)", maxDepth)
    }
    
    var unexpanded []string
    var result strings.Builder
    lastIndex := 0
    
    matches := templateRegex.FindAllStringSubmatchIndex(input, -1)
    if len(matches) == 0 {
        return input, nil, nil
    }
    
    for _, match := range matches {
        result.WriteString(input[lastIndex:match[0]])
        
        fullMatch := input[match[0]:match[1]]
        varName := input[match[2]:match[3]]
        
        if match[0] > 0 && input[match[0]-1] == '$' {
            result.WriteString(fullMatch)
            lastIndex = match[1]
            continue
        }
        
        var defaultValue string
        if match[4] != -1 {
            defaultValue = input[match[4]:match[5]]
        }
        
        envValue := os.Getenv(varName)
        if envValue != "" {
            result.WriteString(envValue)
        } else if defaultValue != "" {
            result.WriteString(defaultValue)
        } else if defaultValue == "" {
            unexpanded = append(unexpanded, varName)
            result.WriteString(fullMatch)
        }
        
        lastIndex = match[1]
    }
    
    result.WriteString(input[lastIndex:])
    output := result.String()
    
    if output != input {
        expanded, moreUnexpanded, err := expandRecursive(output, depth+1)
        if err != nil {
            return "", nil, err
        }
        unexpanded = append(unexpanded, moreUnexpanded...)
        return expanded, unexpanded, nil
    }
    
    return output, unexpanded, nil
}