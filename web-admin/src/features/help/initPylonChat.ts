import type { V1User } from "../../client";

const appId = import.meta.env.RILL_UI_PUBLIC_PYLON_APP_ID as string;

/**
 * Function implementation is copied from: https://docs.usepylon.com/chat/setup
 */
export function initPylonChat(user: V1User) {
  window.pylon = {
    chat_settings: {
      app_id: appId,
      email: user.email,
      name: user.displayName,
      avatar_url: user.photoUrl,
    },
  };
}
