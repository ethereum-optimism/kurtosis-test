sut = import_module("/sut.star")

def test_true(plan):
    expect.true(True)

def test_false(plan):
    expect.true(not False)

def test_fail(plan):
    expect.fails(lambda : sut.sut_fail("oh no"), "oh no")