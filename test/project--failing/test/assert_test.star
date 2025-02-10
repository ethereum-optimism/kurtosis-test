sut = import_module("/sut.star")

def test_assert_true(plan):
    assert.true(sut.sut_return_value(False))

def test_assert_eq(plan):
    assert.eq(sut.sut_return_value(0), sut.sut_return_value(1))

def test_assert_fails(plan):
    assert.fails(lambda : 0, "this did not fail")

def test_multiple_asserts(plan):
    assert.true(False)
    assert.true(not True)