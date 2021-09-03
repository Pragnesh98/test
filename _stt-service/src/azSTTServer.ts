import { server } from './handlers/azureSttHandler'

type StartServerType = () => void;
export const startServer: StartServerType = (): void => {
  server.listen(8081, (err : Error, address: String) => {
    if(err) {
      console.error(err);
      process.exit(0);
    }
    console.log(`\n${new Date().toLocaleString()} Server listening at ${address}\n`);
  });
}