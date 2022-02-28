import type {DataModelerStateService} from "../data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DatabaseService} from "$common/database-service/DatabaseService";
import type { NotificationService } from "$common/notifications/NotificationService";
import { ActionsBase } from "$common/ActionsBase";
import type { RootConfig } from "$common/config/RootConfig";

/**
 * Class that has the actual action implementations.
 */
export class DataModelerActions extends ActionsBase {
    protected dataModelerService: DataModelerService;
    protected notificationService: NotificationService;

    constructor(protected readonly config: RootConfig,
                protected readonly dataModelerStateService: DataModelerStateService,
                protected readonly databaseService: DatabaseService) {
        super();
    }

    public setDataModelerActionService(dataModelerService: DataModelerService): void {
        this.dataModelerService = dataModelerService;
    }

    public setNotificationService(notificationService: NotificationService): void {
        this.notificationService = notificationService;
    }
}
