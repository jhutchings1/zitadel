import { DataSource } from '@angular/cdk/collections';
import { BehaviorSubject, from, Observable, of } from 'rxjs';
import { catchError, finalize, map } from 'rxjs/operators';
import { ProjectMemberView } from 'src/app/proto/generated/management_pb';
import { ProjectService } from 'src/app/services/project.service';

/**
 * Data source for the ProjectMembers view. This class should
 * encapsulate all logic for fetching and manipulating the displayed data
 * (including sorting, pagination, and filtering).
 */
export class ProjectGrantDetailDataSource extends DataSource<ProjectMemberView.AsObject> {
    public totalResult: number = 0;
    public membersSubject: BehaviorSubject<ProjectMemberView.AsObject[]>
        = new BehaviorSubject<ProjectMemberView.AsObject[]>([]);
    private loadingSubject: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(false);
    public loading$: Observable<boolean> = this.loadingSubject.asObservable();

    constructor(private projectService: ProjectService) {
        super();
    }

    public loadMembers(projectId: string, grantId: string,
        pageIndex: number, pageSize: number, sortDirection?: string): void {
        const offset = pageIndex * pageSize;

        this.loadingSubject.next(true);

        from(this.projectService.SearchProjectGrantMembers(projectId, grantId, pageSize, offset)).pipe(
            map(resp => {
                this.totalResult = resp.toObject().totalResult;
                return resp.toObject().resultList;
            }),
            catchError(() => of([])),
            finalize(() => this.loadingSubject.next(false)),
        ).subscribe(members => {
            this.membersSubject.next(members);
        });
    }


    /**
     * Connect this data source to the table. The table will only update when
     * the returned stream emits new items.
     * @returns A stream of the items to be rendered.
     */
    public connect(): Observable<ProjectMemberView.AsObject[]> {
        return this.membersSubject.asObservable();
    };

    /**
     *  Called when the table is being destroyed. Use this function, to clean up
     * any open connections or free any held resources that were set up during connect.
     */
    public disconnect(): void {
        this.membersSubject.complete();
        this.loadingSubject.complete();
    }
}
