import { DefaultAzureCredential } from "@azure/identity"
import { blobClient, containerClient } from "./client"


jest.setTimeout(5000);

describe("client", () => {
    let originalEnv = process.env
    beforeEach(() => {
         process.env['AZURE_TENANT_ID'] = "mock-tenant-id"
         process.env['AZURE_CLIENT_ID'] = "mock-client-id"
         process.env['AZURE_CLIENT_SECRET'] = "mock-secret-id-978wef9w4hf9h2403h04ef"
    })
    test("should successfully initialize blob client", async () =>{
        let cred = new DefaultAzureCredential()
        let client = blobClient(cred, "someAccount")
        expect(client.accountName).toBe("someAccount")
    })
    test("should successfully return a container client", async () =>{
        let cred = new DefaultAzureCredential()
        let client = blobClient(cred, "someAccount")
        let containerCl = containerClient(client, "blobContainer")
        expect(containerCl.containerName).toBe("blobContainer")
    })
    afterEach(() => {
        process.env = { ...originalEnv}
    })
})
