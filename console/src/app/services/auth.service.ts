import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { AuthConfig, OAuthService } from 'angular-oauth2-oidc';
import { BehaviorSubject, from, merge, Observable, of, Subject } from 'rxjs';
import { catchError, filter, finalize, first, map, mergeMap, switchMap, take, timeout } from 'rxjs/operators';

import { Org, UserProfileView } from '../proto/generated/auth_pb';
import { AuthUserService } from './auth-user.service';
import { GrpcService } from './grpc.service';
import { StatehandlerService } from './statehandler.service';
import { StorageKey, StorageService } from './storage.service';

@Injectable({
    providedIn: 'root',
})
export class AuthService {
    private cachedOrgs: Org.AsObject[] = [];
    private _activeOrgChanged: Subject<Org.AsObject> = new Subject();
    public user!: Observable<UserProfileView.AsObject>;
    private _authenticated: boolean = false;
    private readonly _authenticationChanged: BehaviorSubject<
        boolean
    > = new BehaviorSubject(this.authenticated);

    private zitadelPermissions: BehaviorSubject<string[]> = new BehaviorSubject(['user.resourceowner']);

    constructor(
        private grpcService: GrpcService,
        private config: AuthConfig,
        private oauthService: OAuthService,
        private userService: AuthUserService,
        private storage: StorageService,
        private statehandler: StatehandlerService,
        private router: Router,
    ) {
        this.user = merge(
            of(this.oauthService.getAccessToken()).pipe(
                filter(token => token ? true : false),
            ),
            this.oauthService.events.pipe(
                filter(e => e.type === 'token_received'),
                timeout(this.oauthService.waitForTokenInMsec || 0),
                catchError(_ => of(null)), // timeout is not an error
                map(_ => this.oauthService.getAccessToken()),
            ),
        ).pipe(
            take(1),
            mergeMap(token => {
                console.log(token);
                return from(this.userService.GetMyUserProfile().then(userprofile => userprofile.toObject()));
            }),
            finalize(() => {
                this.loadPermissions();
            }),
        );

        this.activeOrgChanged.subscribe(() => {
            console.log('org change');
            this.loadPermissions();
        });
    }

    private loadPermissions(): void {
        console.log('load permissions');
        merge([
            // this.authenticationChanged,
            this.activeOrgChanged.pipe(map(org => !!org)),
        ]).pipe(
            first(),
            switchMap(() => from(this.userService.GetMyzitadelPermissions())),
            map(rolesResp => rolesResp.toObject().permissionsList),
        ).subscribe(roles => this.zitadelPermissions.next(roles));
    }

    public isAllowed(roles: string[], each: boolean = false): Observable<boolean> {
        if (roles && roles.length > 0) {
            return this.zitadelPermissions.pipe(switchMap(zroles => {
                return of(this.hasRoles(zroles, roles, each));
            }));
        } else {
            return of(false);
        }
    }

    public hasRoles(userRoles: string[], requestedRoles: string[], each: boolean = false): boolean {
        // console.log('has', userRoles);
        // console.log('needs', requestedRoles);
        return each ?
            requestedRoles.every(role => userRoles.includes(role)) :
            requestedRoles.findIndex(role => {
                return userRoles.findIndex(i => i.includes(role)) > -1;
            }) > -1;
    }

    public get authenticated(): boolean {
        return this._authenticated;
    }

    public get authenticationChanged(): Observable<boolean> {
        return this._authenticationChanged;
    }

    public getOIDCUser(): Observable<any> {
        return from(this.oauthService.loadUserProfile());
    }

    public async authenticate(
        config?: Partial<AuthConfig>,
        setState: boolean = true,
        force: boolean = false,
    ): Promise<boolean> {
        this.config.issuer = config?.issuer || this.grpcService.issuer;
        this.config.clientId = config?.clientId || this.grpcService.clientid;
        this.config.redirectUri = config?.redirectUri || this.grpcService.redirectUri;
        this.config.postLogoutRedirectUri = config?.postLogoutRedirectUri || this.grpcService.postLogoutRedirectUri;
        this.config.customQueryParams = config?.customQueryParams;
        this.oauthService.configure(this.config);
        // this.oauthService.setupAutomaticSilentRefresh();
        this.oauthService.strictDiscoveryDocumentValidation = false;
        await this.oauthService.loadDiscoveryDocumentAndTryLogin();

        this._authenticated = this.oauthService.hasValidAccessToken();

        if (!this.oauthService.hasValidIdToken() || !this.authenticated || config || force) {
            const newState = setState ? await this.statehandler.createState().toPromise() : undefined;
            this.oauthService.initCodeFlow(newState);
        }
        this._authenticationChanged.next(this.authenticated);

        return this.authenticated;
    }

    public signout(): void {
        this.oauthService.logOut();
        this._authenticated = false;
        this._authenticationChanged.next(false);
        this.router.navigate(['/']);
    }

    public get activeOrgChanged(): Observable<Org.AsObject> {
        return this._activeOrgChanged;
    }

    public async GetActiveOrg(id?: string): Promise<Org.AsObject> {
        if (id) {
            const org = this.storage.getItem<Org.AsObject>(StorageKey.organization);
            if (org && this.cachedOrgs.find(tmp => tmp.id === org.id)) {
                return org;
            }
            return Promise.reject(new Error('no cached org'));
        } else {
            let orgs = this.cachedOrgs;
            if (orgs.length === 0) {
                orgs = (await this.userService.SearchMyProjectOrgs(10, 0)).toObject().resultList;
                this.cachedOrgs = orgs;
            }

            const org = this.storage.getItem<Org.AsObject>(StorageKey.organization);
            if (org && orgs.find(tmp => tmp.id === org.id)) {
                return org;
            }

            if (orgs.length === 0) {
                return Promise.reject(new Error('No organizations found!'));
            }
            const orgToSet = orgs.find(element => element.id !== '0' && element.name !== '');

            if (orgToSet) {
                this.setActiveOrg(orgToSet);
                return Promise.resolve(orgToSet);
            }
            return Promise.resolve(orgs[0]);
        }
    }

    public setActiveOrg(org: Org.AsObject): void {
        this.storage.setItem(StorageKey.organization, org);
        this._activeOrgChanged.next(org);
    }
}

