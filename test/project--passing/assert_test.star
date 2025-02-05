sut = import_module("/sut.star")

def test_true(plan):
    assert.true(True)

def test_false(plan):
    assert.true(not False)

def test_fail(plan):
    assert.fails(lambda : sut.sut_fail("oh no"), "oh no")