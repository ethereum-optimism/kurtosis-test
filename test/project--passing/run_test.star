def test_run_python(plan):
    result = plan.run_python(
        run = """
        print("running")    
        """,
    )

    assert.ne(result, None)

def test_run_sh(plan):
    result = plan.run_sh(
        run = "ls",
    )

    assert.ne(result, None)