import { aiConsoleTickets } from './productionLink';
import { post } from './request';

// AIConsoleTicket is the URL that opens the AI console, holding a ticket
// good once, for a minute.
export interface AIConsoleTicket {
  url: string;
}

// createAIConsoleTicket asks for a ticket that opens a project's AI workspace,
// or the platform's AI monitor. Errors come back as the server's business
// code, for the page to explain.
export function createAIConsoleTicket(params: { project?: string; monitor?: boolean }): Promise<AIConsoleTicket> {
  return post(aiConsoleTickets, params, true);
}
