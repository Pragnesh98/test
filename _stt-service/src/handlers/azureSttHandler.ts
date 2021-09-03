import fastify from "fastify";
import { recognizer as azureRecognizer } from "../utils/recognizeOnce";
import { RecognizeRequest } from "../models/recgonize_request";
import { RecognizeResponse } from "../models/recognize_response";
import { RouteGenericInterface } from "fastify/types/route";

export const server = fastify({
    logger: true,
});

interface RecognizeAPI extends RouteGenericInterface {
    Body: any;
    Reply: RecognizeResponse;
}

server.post<RecognizeAPI>("/azure/recognize",
    async (request, reply) => {
        const response = await azureRecognizer(request.body);
        console.log("!!!!!", response)
        reply.send(response);
    }
);
