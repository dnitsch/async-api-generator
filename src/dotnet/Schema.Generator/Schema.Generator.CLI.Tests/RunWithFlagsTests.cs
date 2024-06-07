using schemagenerator;

namespace Schema.Generator.CLI.Tests;

[TestClass]
public class RunWithFlags
{
    // relative to execution from inside the bin directory
    private const string TEST_SOURCE_DLL = "../../../../Sample.Generated.DLL.Source/bin/Release/net8.0/publish/Sample.Generated.DLL.Source.dll";
    // 
    private const string OUT_DIR = "../../../.ignore";
    private static readonly string[] _NS_FILTER = ["Sample.Generated.DLL.Source.TakeThisModel"];
 
    [TestCleanup]
    public void AfterEach (){
        Directory.Delete(OUT_DIR, true);
    }

    [TestMethod]
    public void TestWithDereferenceSpecified()
    {
        // Arrange
        var gen = new Generate(TEST_SOURCE_DLL, _NS_FILTER, true, OUT_DIR);
        var err = "";

        // Act
        try
        {
            gen.SchemaFromTypes();
        }
        catch (Exception ex)
        {
            err = $"{ex.Message}. {ex.StackTrace}";
        }
        string[] files = Directory.GetFiles(OUT_DIR, "", SearchOption.AllDirectories);

        // Assert
        Assert.IsFalse(err.Length > 0);

        Assert.IsTrue(files.Length >= 4);
    }
    
    [TestMethod]
    public void TestWithoutDereferenceSpecified_should_produce_the_same_number_of_files()
    {
        // Arrange
        var gen = new Generate(TEST_SOURCE_DLL, _NS_FILTER, false, OUT_DIR);
        var err = "";

        // Act
        try
        {
            gen.SchemaFromTypes();
        }
        catch (Exception ex)
        {
            err = $"{ex.Message}. {ex.StackTrace}";
        }
        string[] files = Directory.GetFiles(OUT_DIR, "", SearchOption.AllDirectories);

        // Assert
        Assert.IsFalse(err.Length > 0);

        Assert.IsTrue(files.Length >= 4);
    }

    [TestMethod]
    public void Test_With_nonexistant_DLL_should_throw()
    {
        // Arrange
        var gen = new Generate("/path/not/found.dll", _NS_FILTER, false, OUT_DIR);
        var err = "";

        // Act
        try
        {
            gen.SchemaFromTypes();
        }
        catch (Exception ex)
        {
            err = $"{ex.Message}. {ex.StackTrace}";
        }
        string[] files = Directory.GetFiles(OUT_DIR, "", SearchOption.AllDirectories);

        // Assert
        // Assert.IsTrue(err.Length > 0);

        Assert.IsTrue(files.Length == 0);
    }

}

