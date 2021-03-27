#include<stdio.h>
#include<string.h>

void main()
{
	int i=0,cnt=0;
	char str1[20]={0},str2[20]={0},temp1[20]={0},temp2[20]={0};

	printf("Please Enter CIDR:");
	gets(str1);
	__fpurge(stdin);

	printf("Please put Plain IP here: ");
	gets(str2);
	__fpurge(stdin);

	for(i=strlen(str1);str1[i]!= '.';i--)
	{
		cnt++;
	}
	memmove(temp1,str1,strlen(str1)-cnt);
	printf("temp1=%s,cnt=%d\n",temp1,cnt);
	cnt=0,i=0;
	for(i=strlen(str2);str2[i]!= '.';i--)
	{
		cnt++;
	}
	memmove(temp2,str2,strlen(str2)-cnt);
	printf("temp2=%s,cnt=%d\n",temp2,cnt);

	if(strcmp(temp1,temp2)==0)
		printf("IP matched\n");
	else
		printf("IP not matched\n");

}
