import { OverlayModule } from '@angular/cdk/overlay';
import { CommonModule, registerLocaleData } from '@angular/common';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import localeDe from '@angular/common/locales/de';
import { APP_INITIALIZER, NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ServiceWorkerModule } from '@angular/service-worker';
import { TranslateLoader, TranslateModule } from '@ngx-translate/core';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { AuthConfig, OAuthModule, OAuthStorage } from 'angular-oauth2-oidc';
import { QuicklinkModule } from 'ngx-quicklink';

import { environment } from '../environments/environment';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { HasRoleModule } from './directives/has-role/has-role.module';
import { OutsideClickModule } from './directives/outside-click/outside-click.module';
import { AccountsCardModule } from './modules/accounts-card/accounts-card.module';
import { AvatarModule } from './modules/avatar/avatar.module';
import { WarnDialogModule } from './modules/warn-dialog/warn-dialog.module';
import { SignedoutComponent } from './pages/signedout/signedout.component';
import { HasRolePipeModule } from './pipes/has-role-pipe.module';
import { AuthUserService } from './services/auth-user.service';
import { AuthService } from './services/auth.service';
import { GrpcService } from './services/grpc.service';
import { GrpcAuthInterceptor } from './services/interceptors/grpc-auth.interceptor';
import { GRPC_INTERCEPTORS } from './services/interceptors/grpc-interceptor';
import { GrpcOrgInterceptor } from './services/interceptors/grpc-org.interceptor';
import { StatehandlerProcessorService, StatehandlerProcessorServiceImpl } from './services/statehandler-processor.service';
import { StatehandlerService, StatehandlerServiceImpl } from './services/statehandler.service';
import { StorageService } from './services/storage.service';
import { ThemeService } from './services/theme.service';

registerLocaleData(localeDe);

// AoT requires an exported function for factories
export function HttpLoaderFactory(http: HttpClient): TranslateHttpLoader {
    return new TranslateHttpLoader(http, './assets/i18n/');
}

const appInitializerFn = (grpcServ: GrpcService) => {
    return () => {
        return grpcServ.loadAppEnvironment();
    };
};

const stateHandlerFn = (stateHandler: StatehandlerService) => {
    return () => {
        return stateHandler.initStateHandler();
    };
};

const authConfig: AuthConfig = {
    scope: 'openid profile email', // offline_access
    responseType: 'code',
    oidc: true,
};

@NgModule({
    declarations: [
        AppComponent,
        SignedoutComponent,
    ],
    imports: [
        AppRoutingModule,
        CommonModule,
        BrowserModule,
        OverlayModule,
        OAuthModule.forRoot({
            resourceServer: {
                allowedUrls: ['https://test.api.zitadel.caos.ch/caos.zitadel.auth.api.v1.AuthService', 'https://test.api.zitadel.caos.ch/oauth/v2/userinfo', 'https://test.api.zitadel.caos.ch/caos.zitadel.management.api.v1.ManagementService/', 'https://preview.api.zitadel.caos.ch'],
                sendAccessToken: true,
            },
        }),
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: HttpLoaderFactory,
                deps: [HttpClient],
            },
        }),
        QuicklinkModule,
        AccountsCardModule,
        HasRoleModule,
        BrowserAnimationsModule,
        HttpClientModule,
        MatButtonModule,
        MatIconModule,
        MatTooltipModule,
        MatSidenavModule,
        MatCardModule,
        OutsideClickModule,
        HasRolePipeModule,
        MatProgressBarModule,
        MatToolbarModule,
        MatMenuModule,
        MatSnackBarModule,
        AvatarModule,
        WarnDialogModule,
        MatDialogModule,
        ServiceWorkerModule.register('ngsw-worker.js', { enabled: environment.production }),
    ],
    providers: [
        ThemeService,
        {
            provide: APP_INITIALIZER,
            useFactory: appInitializerFn,
            multi: true,
            deps: [GrpcService],
        },
        {
            provide: APP_INITIALIZER,
            useFactory: stateHandlerFn,
            multi: true,
            deps: [StatehandlerService],
        },
        {
            provide: AuthConfig,
            useValue: authConfig,
        },
        {
            provide: StatehandlerProcessorService,
            useClass: StatehandlerProcessorServiceImpl,
        },
        {
            provide: StatehandlerService,
            useClass: StatehandlerServiceImpl,
        },
        {
            provide: OAuthStorage,
            useClass: StorageService,
        },
        {
            provide: GRPC_INTERCEPTORS,
            multi: true,
            useClass: GrpcAuthInterceptor,
        },
        {
            provide: GRPC_INTERCEPTORS,
            multi: true,
            useClass: GrpcOrgInterceptor,
        },
        GrpcService,
        AuthService,
        AuthUserService,
        { provide: 'windowObject', useValue: window },
    ],
    bootstrap: [AppComponent],
})
export class AppModule {

    constructor() { }
}
