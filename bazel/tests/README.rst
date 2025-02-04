gomock
=====================

Tests that ensure the gomock rules can be called correctly under different input permutations.

reflective
------------------------
Checks that gomock can be run in "reflective" mode when passed a `GoLibrary` and `interfaces`.

source
------------------------
Checks that gomock can be run in "source" mode when passed a `GoLibrary` and `source`.

source_with_importpath
------------------------
Checks that gomock can be run in "source" mode when passed an `importpath` and `source`.
This test case also demonstrates the circumstance in which `importpath` is necessary to prevent a circular dependency.
