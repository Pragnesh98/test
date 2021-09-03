import 'dotenv/config';
import * as grpc from 'grpc';
import sttHandler from './handlers/streamSttHandler';

const port: string | number = process.env.PORT || 50051;

type StartServerType = () => void;
export const startServer: StartServerType = (): void => {
    const server: grpc.Server = new grpc.Server();

    // register all the handler here...
    server.addService(sttHandler.service, sttHandler.handler);

    // define the host/port for server
    server.bindAsync(
        `0.0.0.0:${ port }`,
        grpc.ServerCredentials.createInsecure(),
        (err: Error|null, port: number) => {
            if (err != null) {
                return console.error(err);
            }
            console.log(`\ngRPC listening on ${ port }\n`);
        },
    );

    // start the gRPC server
    server.start();
};
