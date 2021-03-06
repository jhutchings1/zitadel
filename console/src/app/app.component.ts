import { BreakpointObserver } from '@angular/cdk/layout';
import { OverlayContainer } from '@angular/cdk/overlay';
import { ViewportScroller } from '@angular/common';
import { Component, HostBinding, Inject, OnDestroy, ViewChild } from '@angular/core';
import { MatIconRegistry } from '@angular/material/icon';
import { MatDrawer } from '@angular/material/sidenav';
import { DomSanitizer } from '@angular/platform-browser';
import { Router, RouterOutlet } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { Observable, of, Subscription } from 'rxjs';
import { map } from 'rxjs/operators';

import { accountCard, navAnimations, routeAnimations, toolbarAnimation } from './animations';
import { Org, UserProfileView } from './proto/generated/auth_pb';
import { AuthUserService } from './services/auth-user.service';
import { AuthService } from './services/auth.service';
import { ProjectService } from './services/project.service';
import { ThemeService } from './services/theme.service';
import { ToastService } from './services/toast.service';
import { UpdateService } from './services/update.service';

@Component({
    selector: 'app-root',
    templateUrl: './app.component.html',
    styleUrls: ['./app.component.scss'],
    animations: [
        toolbarAnimation,
        ...navAnimations,
        accountCard,
        routeAnimations,
    ],
})
export class AppComponent implements OnDestroy {
    @ViewChild('drawer')
    public drawer!: MatDrawer;
    public isHandset$: Observable<boolean> = this.breakpointObserver
        .observe('(max-width: 599px)')
        .pipe(map(result => {
            return result.matches;
        }));
    @HostBinding('class') public componentCssClass: string = 'dark-theme';

    public showAccount: boolean = false;
    public org!: Org.AsObject;
    public orgs: Org.AsObject[] = [];
    public profile!: UserProfileView.AsObject;
    public isDarkTheme: Observable<boolean> = of(true);

    public orgLoading: boolean = false;

    public showProjectSection: boolean = false;

    public grantedProjectsCount: number = 0;
    public ownedProjectsCount: number = 0;

    private authSub: Subscription = new Subscription();
    private orgSub: Subscription = new Subscription();

    constructor(
        public viewPortScroller: ViewportScroller,
        @Inject('windowObject') public window: Window,
        public translate: TranslateService,
        public authService: AuthService,
        private breakpointObserver: BreakpointObserver,
        public overlayContainer: OverlayContainer,
        private themeService: ThemeService,
        public userService: AuthUserService,
        private projectService: ProjectService,
        public matIconRegistry: MatIconRegistry,
        public domSanitizer: DomSanitizer,
        private toast: ToastService,
        private router: Router,
        update: UpdateService,
    ) {
        console.log('%cWait!', 'text-shadow: -1px 0 black, 0 1px black, 1px 0 black, 0 -1px black; color: #5282c1; font-size: 50px');
        console.log('%cInserting something here could give attackers access to your zitadel account.', 'color: red; font-size: 18px');
        console.log('%cIf you don\'t know exactly what you\'re doing, close the window and stay on the safe side', 'font-size: 16px');
        console.log('%cIf you know exactly what you are doing, you should work for us', 'font-size: 16px');
        this.setLanguage();

        this.matIconRegistry.addSvgIcon(
            'mdi_account_check_outline',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/account-check-outline.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_account_cancel',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/account-cancel-outline.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_light_on',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/lightbulb-on-outline.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_content_copy',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/content-copy.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_light_off',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/lightbulb-off-outline.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_radar',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/radar.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_lock_question',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/lock-question.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_textbox_password',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/textbox-password.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_lock_reset',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/lock-reset.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_broom',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/broom.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_pin_outline',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/pin-outline.svg'),
        );

        this.matIconRegistry.addSvgIcon(
            'mdi_pin',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/mdi/pin.svg'),
        );
        this.getProjectCount();

        this.orgSub = this.authService.activeOrgChanged.subscribe(org => {
            this.org = org;

            this.getProjectCount();
        });

        this.authSub = this.authService.authenticationChanged.subscribe((authenticated) => {
            if (authenticated) {
                this.authService.GetActiveOrg().then(org => {
                    this.org = org;
                });
            }
        });

        const theme = localStorage.getItem('theme');
        if (theme) {
            this.overlayContainer.getContainerElement().classList.add(theme);
            this.componentCssClass = theme;
        }

        this.isDarkTheme = this.themeService.isDarkTheme;
        this.isDarkTheme.subscribe(thema => this.onSetTheme(thema ? 'dark-theme' : 'light-theme'));
    }

    public ngOnDestroy(): void {
        this.authSub.unsubscribe();
        this.orgSub.unsubscribe();
    }

    public loadOrgs(): void {
        this.orgLoading = true;
        this.userService.SearchMyProjectOrgs(10, 0).then(res => {
            this.orgs = res.toObject().resultList;
            this.orgLoading = false;
        }).catch(error => {
            this.toast.showError(error);
            this.orgLoading = false;
        });
    }

    public prepareRoute(outlet: RouterOutlet): boolean {
        return outlet && outlet.activatedRouteData && outlet.activatedRouteData.animation;
    }

    public closeAccountCard(): void {
        if (this.showAccount) {
            this.showAccount = false;
        }
    }

    public onSetTheme(theme: string): void {
        localStorage.setItem('theme', theme);
        this.overlayContainer.getContainerElement().classList.add(theme);
        this.componentCssClass = theme;
    }

    private setLanguage(): void {
        this.translate.addLangs(['en', 'de']);
        this.translate.setDefaultLang('en');

        this.authService.user.subscribe(userprofile => {
            this.profile = userprofile;
            const lang = userprofile.preferredLanguage.match(/en|de/) ? userprofile.preferredLanguage : 'en';
            this.translate.use(lang);
        });
    }

    public setActiveOrg(org: Org.AsObject): void {
        this.org = org;
        this.authService.setActiveOrg(org);
        this.router.navigate(['/']);
    }

    private async getProjectCount(): Promise<any> {
        this.ownedProjectsCount = await this.projectService.SearchProjects(0, 0).then(res => {
            return res.toObject().totalResult;
        });

        this.grantedProjectsCount = await this.projectService.SearchGrantedProjects(0, 0).then(res => {
            return res.toObject().totalResult;
        });
    }
}

