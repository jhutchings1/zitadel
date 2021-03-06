import { SelectionModel } from '@angular/cdk/collections';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { MatButtonToggleChange } from '@angular/material/button-toggle';
import { MatDialog } from '@angular/material/dialog';
import { MatTableDataSource } from '@angular/material/table';
import { Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { BehaviorSubject, from, Observable, of, Subscription } from 'rxjs';
import { catchError, finalize, map } from 'rxjs/operators';
import { CreationType, MemberCreateDialogComponent } from 'src/app/modules/add-member-dialog/member-create-dialog.component';
import { ChangeType } from 'src/app/modules/changes/changes.component';
import { WarnDialogComponent } from 'src/app/modules/warn-dialog/warn-dialog.component';
import {
    Org,
    OrgDomainView,
    OrgMember,
    OrgMemberSearchResponse,
    OrgMemberView,
    OrgState,
    User,
} from 'src/app/proto/generated/management_pb';
import { OrgService } from 'src/app/services/org.service';
import { ToastService } from 'src/app/services/toast.service';

import { AddDomainDialogComponent } from './add-domain-dialog/add-domain-dialog.component';


@Component({
    selector: 'app-org-detail',
    templateUrl: './org-detail.component.html',
    styleUrls: ['./org-detail.component.scss'],
})
export class OrgDetailComponent implements OnInit, OnDestroy {
    public org!: Org.AsObject;

    public dataSource: MatTableDataSource<OrgMember.AsObject> = new MatTableDataSource<OrgMember.AsObject>();
    public memberResult!: OrgMemberSearchResponse.AsObject;
    public displayedColumns: string[] = ['select', 'firstname', 'lastname', 'username', 'email', 'roles'];
    public selection: SelectionModel<OrgMember.AsObject> = new SelectionModel<OrgMember.AsObject>(true, []);
    public OrgState: any = OrgState;
    public ChangeType: any = ChangeType;

    private subscription: Subscription = new Subscription();

    public domains: OrgDomainView.AsObject[] = [];
    public primaryDomain: string = '';

    // members
    private loadingSubject: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(false);
    public loading$: Observable<boolean> = this.loadingSubject.asObservable();
    public totalMemberResult: number = 0;
    public membersSubject: BehaviorSubject<OrgMemberView.AsObject[]>
        = new BehaviorSubject<OrgMemberView.AsObject[]>([]);

    constructor(
        private dialog: MatDialog,
        public translate: TranslateService,
        private orgService: OrgService,
        private toast: ToastService,
        private router: Router,
    ) { }

    public ngOnInit(): void {
        this.getData();
    }

    public ngOnDestroy(): void {
        this.subscription.unsubscribe();
    }

    private async getData(): Promise<void> {
        this.orgService.GetMyOrg().then((org: Org) => {
            this.org = org.toObject();
        }).catch(error => {
            this.toast.showError(error);
        });

        this.loadingSubject.next(true);
        from(this.orgService.SearchMyOrgMembers(100, 0)).pipe(
            map(resp => {
                this.totalMemberResult = resp.toObject().totalResult;
                return resp.toObject().resultList;
            }),
            catchError(() => of([])),
            finalize(() => this.loadingSubject.next(false)),
        ).subscribe(members => {
            this.membersSubject.next(members);
        });

        this.orgService.SearchMyOrgDomains(0, 100).then(result => {
            this.domains = result.toObject().resultList;
            this.primaryDomain = this.domains.find(domain => domain.primary)?.domain ?? '';
        });
    }

    public changeState(event: MatButtonToggleChange | any): void {
        if (event.value === OrgState.ORGSTATE_ACTIVE) {
            this.orgService.ReactivateMyOrg().then(() => {
                this.toast.showInfo('ORG.TOAST.REACTIVATED', true);
            }).catch((error) => {
                this.toast.showError(error);
            });
        } else if (event.value === OrgState.ORGSTATE_INACTIVE) {
            this.orgService.DeactivateMyOrg().then(() => {
                this.toast.showInfo('ORG.TOAST.DEACTIVATED', true);
            }).catch((error) => {
                this.toast.showError(error);
            });
        }
    }

    public addNewDomain(): void {
        const dialogRef = this.dialog.open(AddDomainDialogComponent, {
            data: {},
            width: '400px',
        });

        dialogRef.afterClosed().subscribe(resp => {
            if (resp) {
                this.orgService.AddMyOrgDomain(resp).then(domain => {
                    this.domains.push(domain.toObject());
                    this.toast.showInfo('ORG.TOAST.DOMAINADDED', true);
                });
            }
        });
    }

    public removeDomain(domain: string): void {
        const dialogRef = this.dialog.open(WarnDialogComponent, {
            data: {
                confirmKey: 'ACTIONS.DELETE',
                cancelKey: 'ACTIONS.CANCEL',
                titleKey: 'ORG.DOMAINS.DELETE.TITLE',
                descriptionKey: 'ORG.DOMAINS.DELETE.DESCRIPTION',
            },
            width: '400px',
        });

        dialogRef.afterClosed().subscribe(resp => {
            if (resp) {
                this.orgService.RemoveMyOrgDomain(domain).then(() => {
                    this.toast.showInfo('ORG.TOAST.DOMAINREMOVED', true);
                    const index = this.domains.findIndex(d => d.domain === domain);
                    if (index > -1) {
                        this.domains.splice(index, 1);
                    }
                }).catch(error => {
                    this.toast.showError(error);
                });
            }
        });
    }

    public openAddMember(): void {
        const dialogRef = this.dialog.open(MemberCreateDialogComponent, {
            data: {
                creationType: CreationType.ORG,
            },
            width: '400px',
        });

        dialogRef.afterClosed().subscribe(resp => {
            if (resp) {
                const users: User.AsObject[] = resp.users;
                const roles: string[] = resp.roles;

                if (users && users.length && roles && roles.length) {
                    Promise.all(users.map(user => {
                        return this.orgService.AddMyOrgMember(user.id, roles);
                    })).then(() => {
                        this.toast.showInfo('ORG.TOAST.MEMBERADDED', true);
                    }).catch(error => {
                        this.toast.showError(error);
                    });
                }
            }
        });
    }

    public showDetail(): void {
        if (this.org?.state === OrgState.ORGSTATE_ACTIVE) {
            this.router.navigate(['org/members']);
        }
    }
}
