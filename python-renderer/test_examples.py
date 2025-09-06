#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Test script for Python CML renderer
This script tests the renderer with all example files
"""

import os
import sys
import glob

# Add current directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from cml_parser import parse_cml_file
from cml_renderer import CMLRenderer


def test_renderer():
    """Test the renderer with all example files."""
    print("Testing Python CML renderer...")
    
    # Create test output directory
    test_output_dir = "test-output"
    if not os.path.exists(test_output_dir):
        os.makedirs(test_output_dir)
    
    # Find all example files
    examples_dir = "../examples"
    example_files = glob.glob(os.path.join(examples_dir, "*.cml"))
    
    if not example_files:
        print("No example files found in ../examples/")
        return False
    
    print("Found {} example files".format(len(example_files)))
    
    success_count = 0
    total_count = len(example_files)
    
    for example_file in example_files:
        example_name = os.path.splitext(os.path.basename(example_file))[0]
        output_file = os.path.join(test_output_dir, "{}.png".format(example_name))
        
        print("Testing: {}".format(example_name))
        
        try:
            # Parse the CML file
            chart = parse_cml_file(example_file)
            
            # Render the chart
            renderer = CMLRenderer()
            renderer.render(chart, output_file)
            
            # Check if output file was created
            if os.path.exists(output_file):
                print("  [OK] Successfully rendered {}".format(example_name))
                success_count += 1
            else:
                print("  [FAIL] Output file not created for {}".format(example_name))
                
        except Exception as e:
            print("  [FAIL] Failed to render {}: {}".format(example_name, e))
            return False
    
    print("")
    print("Test Results:")
    print("  Total examples: {}".format(total_count))
    print("  Successful renders: {}".format(success_count))
    print("  Failed renders: {}".format(total_count - success_count))
    
    if success_count == total_count:
        print("  [OK] All tests passed!")
        print("")
        print("Generated files:")
        for file in glob.glob(os.path.join(test_output_dir, "*.png")):
            print("  {}".format(file))
        return True
    else:
        print("  [FAIL] Some tests failed!")
        return False


if __name__ == "__main__":
    success = test_renderer()
    sys.exit(0 if success else 1)
