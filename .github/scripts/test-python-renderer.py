#!/usr/bin/env python3
import sys
import os
sys.path.append('.')
from cml_parser import parse_cml_file
from cml_renderer import CMLRenderer

def test_render(example_file):
    try:
        chart = parse_cml_file(example_file)
        renderer = CMLRenderer()
        output_file = f'test-output/{os.path.splitext(os.path.basename(example_file))[0]}.png'
        renderer.render(chart, output_file)
        print(f'Successfully rendered {output_file}')
        return True
    except Exception as e:
        print(f'Failed to render {example_file}: {e}')
        return False

if __name__ == '__main__':
    example_file = sys.argv[1]
    success = test_render(example_file)
    sys.exit(0 if success else 1)
