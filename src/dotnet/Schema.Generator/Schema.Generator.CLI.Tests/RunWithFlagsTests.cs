using System.Diagnostics;
using schemagenerator;

namespace Schema.Generator.CLI.Tests;

[TestClass]
public class RunWithFlags
{
    // relative to execution from inside the bin directory
    private const string TEST_SOURCE_DLL = "../../../../Sample.Generated.DLL.Source/bin/Release/net8.0/publish/Sample.Generated.DLL.Source.dll";
    // 
    private const string OUT_DIR = "../../../.ignore";
    private static string[] _NS_FILTER = ["Sample.Generated.DLL.Source.TakeThisModel"];

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
}

