#!/bin/bash
export PROJECT_ID=common

{
    for f in $(git ls-files -- ':!*.jpg' ':!*.png' ':!*.ico' ':!*.svg' ':!assets/*'); do
        echo "// $f"
        cat "app/$f" 2>/dev/null || cat "$f" 2>/dev/null
        echo
        echo "----------------------------------------------"
        echo
    done
} > "../${PROJECT_ID}.txt"

echo "All code copied to ../${PROJECT_ID}.txt"


