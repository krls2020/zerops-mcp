#!/bin/bash

# Script to analyze Zerops recipes
RECIPES_DIR="/Users/macbook/Sites/zerops-mcp-3/docs/zerops-recipes-library"
OUTPUT_FILE="/Users/macbook/Sites/zerops-mcp-3/docs/archive/RECIPES_INVENTORY.md"

# Function to check files in a recipe directory
check_recipe() {
    local recipe_dir="$1"
    local recipe_name=$(basename "$recipe_dir")
    
    echo "Checking: $recipe_name"
    
    # Check for key files
    local has_zerops_yml="no"
    local has_import_yml="no"
    local import_files=""
    local has_readme="no"
    local uses_build_from_git="no"
    
    # Check zerops.yml
    if [ -f "$recipe_dir/zerops.yml" ]; then
        has_zerops_yml="yes"
        # Check if it uses buildFromGit
        if grep -q "buildFromGit:" "$recipe_dir/zerops.yml" 2>/dev/null; then
            uses_build_from_git="yes"
        fi
    fi
    
    # Check for import files
    for file in "$recipe_dir"/*.yml "$recipe_dir"/*.yaml; do
        if [ -f "$file" ] && [[ $(basename "$file") == *"import"* || $(basename "$file") == "project.yml" ]]; then
            has_import_yml="yes"
            import_files="$import_files $(basename "$file")"
        fi
    done
    
    # Check README
    if [ -f "$recipe_dir/README.md" ] || [ -f "$recipe_dir/readme.md" ]; then
        has_readme="yes"
    fi
    
    echo "$recipe_name|$has_zerops_yml|$has_import_yml|$import_files|$has_readme|$uses_build_from_git"
}

# Main execution
echo "# Zerops Recipes Inventory" > "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Generated on: $(date)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Process each recipe
echo "## Raw Data" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "| Recipe Name | Has zerops.yml | Has Import File | Import Files | Has README | Uses buildFromGit |" >> "$OUTPUT_FILE"
echo "|-------------|----------------|-----------------|--------------|------------|-------------------|" >> "$OUTPUT_FILE"

for recipe_dir in "$RECIPES_DIR"/*; do
    if [ -d "$recipe_dir" ]; then
        result=$(check_recipe "$recipe_dir")
        echo "| $result |" >> "$OUTPUT_FILE"
    fi
done

echo "Recipe analysis complete!"